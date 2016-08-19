package docx

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// ReplaceDocx represents a replacable docx
type ReplaceDocx struct {
	zipReader *zip.ReadCloser
	content   string
}

// Editable returns a Docx
func (r *ReplaceDocx) Editable() *Docx {
	return &Docx{
		files:   r.zipReader.File,
		content: r.content,
	}
}

// Close closes the zip reader
func (r *ReplaceDocx) Close() error {
	return r.zipReader.Close()
}

// Docx represents a docx
type Docx struct {
	files   []*zip.File
	content string
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
	d.content = strings.Replace(d.content, oldString, newString, num)

	return nil
}

// ReplaceLoop iterates through the loop in docx
// for each elemen tin the given data array.
// During each run of the iteration, the loop placeholders are replaces with
// the given values in the corresponding data element.
func (d *Docx) ReplaceLoop(loopvar string, data []map[string]string) (err error) {
	var newContent = ""
	buf := bytes.NewBufferString(newContent)

	decoder := xml.NewDecoder(strings.NewReader(d.content))
	encoder := xml.NewEncoder(buf)

	isBefore := true
	isIn := false
	isAfter := true

	var beforeTokens []xml.Token
	var inTokens []xml.Token
	var afterTokens []xml.Token

	// Parse document xml and group tokens in before-loop, part-of-loop, after-loop
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch t.(type) {
		case xml.CharData:
			if charData, ok := t.(xml.CharData); ok {
				cd := strings.Trim(string([]byte(charData)), " ")
				if cd == "«start:"+loopvar+"»" {
					isIn = true
					isBefore = false
					isAfter = false
					continue
				} else if cd == "«end:"+loopvar+"»" {
					isIn = false
					isBefore = false
					isAfter = true
					continue
				} else {
				}
			}
		default:
		}

		if isBefore {
			beforeTokens = append(beforeTokens, xml.CopyToken(t))
		} else if isIn {
			inTokens = append(inTokens, xml.CopyToken(t))
		} else if isAfter {
			afterTokens = append(afterTokens, xml.CopyToken(t))
		}

	}

	// Copy the tokens before the loop (header)
	for _, token := range beforeTokens {
		encoder.EncodeToken(token)
	}

	// Iterate the loop tokens for each loop element
	fmt.Println("Starting parsing loop ...")
	for idx, loopElement := range data {
		fmt.Printf("Iteration %d ... \n", idx)

		// Check if tokens have to be replaced with given data values
		for _, token := range inTokens {
			newToken := xml.CopyToken(token) // TODO: Do we need a copy here?
			switch newToken.(type) {
			case xml.CharData:
				if charData, ok := newToken.(xml.CharData); ok {
					cd := strings.Trim(string([]byte(charData)), " ")
					// TODO: use convention, key in data = <<key>> in docx ?
					if cd == "«name»" {
						newToken = xml.CharData(loopElement["name"])
					} else if cd == "«pos»" {
						newToken = xml.CharData(loopElement["pos"])
					} else if cd == "«user»" {
						newToken = xml.CharData(loopElement["user"])
					}
				}
			default:
			}

			encoder.EncodeToken(newToken)
		}
	}

	// Copy the tokens after the loop
	for _, token := range afterTokens {
		encoder.EncodeToken(token)
	}
	encoder.Flush()

	d.content = buf.String()
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
			writer.Write([]byte(d.content))
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
