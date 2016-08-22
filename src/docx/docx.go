package docx

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"xml"
)

const mergeFieldOpenTag = "«"
const mergeFieldCloseTag = "»"
const loopStartPrefix = "start:"
const loopEndPrefix = "end:"

// ReplaceDocx represents a replacable docx
type ReplaceDocx struct {
	zipReader *zip.ReadCloser
	content   string
}

// Editable returns a Docx
func (r *ReplaceDocx) Editable() *Docx {
	return &Docx{
		files:   r.zipReader.File,
		Content: r.content,
	}
}

// Close closes the zip reader
func (r *ReplaceDocx) Close() error {
	return r.zipReader.Close()
}

// Docx represents a docx
type Docx struct {
	files   []*zip.File
	Content string
}

// Replace replaces a string
func (d *Docx) Replace(oldString string, newString string, num int) (err error) {
	oldString, err = encode(oldString)
	if err != nil {
		return err
	}
	newString, err = encode(newString)
	if err != nil {
		return err
	}

	d.Content = strings.Replace(d.Content, mergeFieldOpenTag+oldString+mergeFieldCloseTag, newString, num)
	return nil
}

// ReplaceLoop iterates through the loop in docx
// for each elemen tin the given data array.
// During each run of the iteration, the loop placeholders are replaces with
// the given values in the corresponding data element.
func (d *Docx) ReplaceLoop(loopVarName string, data []map[string]string) (err error) {
	newContent := ""
	newBuffer := bytes.NewBufferString(newContent)
	newTokens := make(map[string][]xml.Token)
	decoder := xml.NewDecoder(strings.NewReader(d.Content))
	encoder := xml.NewEncoder(newBuffer)
	encoder.OptimizeNamespaces(true)
	encoder.PrefixElements(true)
	encoder.Namespace("w", "http://schemas.openxmlformats.org/wordprocessingml/2006/main")
	encoder.Namespace("w14", "http://schemas.microsoft.com/office/word/2010/wordml")
	encoder.Namespace("xmlns", "xmlns")
	encoder.Namespace("wpc", "http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas")
	encoder.Namespace("mo", "http://schemas.microsoft.com/office/mac/office/2008/main")
	encoder.Namespace("mc", "http://schemas.openxmlformats.org/markup-compatibility/2006")
	encoder.Namespace("mv", "urn:schemas-microsoft-com:mac:vml")
	encoder.Namespace("o", "urn:schemas-microsoft-com:office:office")
	encoder.Namespace("r", "http://schemas.openxmlformats.org/officeDocument/2006/relationships")
	encoder.Namespace("m", "http://schemas.openxmlformats.org/officeDocument/2006/math")
	encoder.Namespace("v", "urn:schemas-microsoft-com:vml")
	encoder.Namespace("wp14", "http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing")
	encoder.Namespace("wp", "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing")
	encoder.Namespace("w10", "urn:schemas-microsoft-com:office:word")
	encoder.Namespace("w", "http://schemas.openxmlformats.org/wordprocessingml/2006/main")
	encoder.Namespace("w14", "http://schemas.microsoft.com/office/word/2010/wordml")
	encoder.Namespace("w15", "http://schemas.microsoft.com/office/word/2012/wordml")
	encoder.Namespace("wpg", "http://schemas.microsoft.com/office/word/2010/wordprocessingGroup")
	encoder.Namespace("wpi", "http://schemas.microsoft.com/office/word/2010/wordprocessingInk")
	encoder.Namespace("wne", "http://schemas.microsoft.com/office/word/2006/wordml")
	encoder.Namespace("wps", "http://schemas.microsoft.com/office/word/2010/wordprocessingShape")

	// pos indicates the position of the token
	// "before" ... token is before the loop block
	// "in"     ... token is part of the loop block
	// "after"  ... token is after of th loop block
	pos := "before"

	// Parse document XML and partition tokens
	// in 3 buckets: before-loop, in-loop, after-loop
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch node := t.(type) {
		case xml.CharData:
			if charData, ok := t.(xml.CharData); ok {
				cd := strings.Trim(string([]byte(charData)), " ")
				if cd == mergeFieldOpenTag+loopStartPrefix+loopVarName+mergeFieldCloseTag {
					pos = "in"
					continue
				} else if cd == mergeFieldOpenTag+loopEndPrefix+loopVarName+mergeFieldCloseTag {
					pos = "after"
					continue
				}
			}
		// WORKAROUND: The genereated document.xml is not fully valid
		// however, MS Word manages to open it with a warning.
		// In order to circumvent the warning, we have to skip the attribute "Ignorable"
		// In the <document/> root element.
		case xml.StartElement:
			if node.Name.Local == "document" {
				var newAttr []xml.Attr
				for _, attr := range node.Attr {
					if strings.ToLower(attr.Name.Local) != "ignorable" {
						newAttr = append(newAttr, attr)
					}
				}
				node.Attr = newAttr
				t = node
			}
		}

		newTokens[pos] = append(newTokens[pos], xml.CopyToken(t))
	}

	// Copy the tokens before the loop (header)
	for _, token := range newTokens["before"] {
		encoder.EncodeToken(token)
	}

	// Iterate the loop tokens for each loop element
	for _, loopElement := range data {
		// Check if tokens have to be replaced with given data values
		for _, token := range newTokens["in"] {
			newToken := xml.CopyToken(token) // TODO: Do we need a copy here?
			switch newToken.(type) {
			case xml.CharData:
				if charData, ok := newToken.(xml.CharData); ok {
					cd := strings.Trim(string([]byte(charData)), " ")
					if strings.HasPrefix(cd, "«") && strings.HasSuffix(cd, "»") {
						key := cd[2 : len(cd)-2]
						if val, ok := loopElement[key]; ok {
							newToken = xml.CharData(val)
						}
					}
				}
			}
			encoder.EncodeToken(newToken)
		}
	}

	// Copy the tokens after the loop
	for _, token := range newTokens["after"] {
		encoder.EncodeToken(token)
	}

	encoder.Flush()

	d.Content = newBuffer.String()
	return nil
}

// WriteToFile writes to file
func (d *Docx) WriteToFile(path string) (err error) {
	var target *os.File
	target, err = os.Create(path)
	if err != nil {
		return
	}
	defer target.Close()
	err = d.Write(target)
	return
}

func (d *Docx) Write(ioWriter io.Writer) (err error) {
	w := zip.NewWriter(ioWriter)
	for _, file := range d.files {
		var writer io.Writer
		var readCloser io.ReadCloser

		writer, err = w.Create(file.Name)
		if err != nil {
			return err
		}
		readCloser, err = file.Open()
		if err != nil {
			return err
		}
		if file.Name == "word/document.xml" {
			writer.Write([]byte(d.Content))
		} else {
			writer.Write(streamToByte(readCloser))
		}
	}
	w.Close()
	return
}

// ReadDocxFile reads the file
func ReadDocxFile(path string) (*ReplaceDocx, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	content, err := readText(reader.File)
	if err != nil {
		return nil, err
	}

	return &ReplaceDocx{zipReader: reader, content: content}, nil
}

// ReadDoxFileFromBytes ...
func ReadDoxFileFromBytes(zipBytes []byte) (*ReplaceDocx, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, err
	}
	content, err := readText(reader.File)
	if err != nil {
		return nil, err
	}

	readCloser := &zip.ReadCloser{Reader: *reader}
	return &ReplaceDocx{zipReader: readCloser, content: content}, nil
}

func readText(files []*zip.File) (text string, err error) {
	var documentFile *zip.File
	documentFile, err = retrieveWordDoc(files)
	if err != nil {
		return text, err
	}
	var documentReader io.ReadCloser
	documentReader, err = documentFile.Open()
	if err != nil {
		return text, err
	}

	text, err = wordDocToString(documentReader)
	return
}

func wordDocToString(reader io.Reader) (string, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func retrieveWordDoc(files []*zip.File) (file *zip.File, err error) {
	for _, f := range files {
		if f.Name == "word/document.xml" {
			file = f
		}
	}
	if file == nil {
		err = errors.New("document.xml file not found")
	}
	return
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func encode(s string) (string, error) {
	var b bytes.Buffer
	enc := xml.NewEncoder(bufio.NewWriter(&b))
	if err := enc.Encode(s); err != nil {
		return s, err
	}
	return strings.Replace(strings.Replace(b.String(), "<string>", "", 1), "</string>", "", 1), nil
}
