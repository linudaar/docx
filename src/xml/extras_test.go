package xml

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SomeStruct struct {
	XMLName Name `xml:"SomeStruct"`
	Name    string
}

type SomeStructWithNS struct {
	XMLName Name `xml:"space SomeStruct"`
	Name    string
}

type SomeStructWithNSFields struct {
	Name string `xml:"space Name"`
}

type RequiredFieldStruct struct {
	RequiredField *SomeStruct `xml:",required"`
}

type NamePrecendeStruct struct {
	NestedStruct *SomeStruct `xml:"NestedStruct,!"`
}

type NamePrecendeWithNSStruct struct {
	NestedStruct *SomeStructWithNS `xml:"space1 NestedStruct,!"`
}

var extraMarshalTests = []struct {
	Value         interface{}
	ExpectXML     string
	MarshalOnly   bool
	UnmarshalOnly bool
}{
	// Test marshalling nil of a required element will result in an empty start tag with xsi:nil=true
	{
		Value:       RequiredFieldStruct{},
		ExpectXML:   `<RequiredFieldStruct><SomeStruct xmlns:xi="http://www.w3.org/2001/XMLSchema-instance" xi:nil="true" /></RequiredFieldStruct>`,
		MarshalOnly: true,
	},
	// Test marshaling with conflicting XML names
	{
		Value:       NamePrecendeStruct{NestedStruct: &SomeStruct{Name: "John Doe"}},
		ExpectXML:   `<NamePrecendeStruct><NestedStruct><Name>John Doe</Name></NestedStruct></NamePrecendeStruct>`,
		MarshalOnly: true,
	},
	// Test marshaling with conflicting XML names and namespace interhitance
	{
		Value:       NamePrecendeWithNSStruct{NestedStruct: &SomeStructWithNS{Name: "John Doe"}},
		ExpectXML:   `<NamePrecendeWithNSStruct><NestedStruct xmlns="space1"><Name>John Doe</Name></NestedStruct></NamePrecendeWithNSStruct>`,
		MarshalOnly: true,
	},
}

func TestMarshalExtras(t *testing.T) {
	for idx, test := range extraMarshalTests {
		data, err := Marshal(test.Value)
		if err != nil {
			t.Errorf("#%d: marshal(%#v): %s", idx, test.Value, err)
			continue
		}
		if got, want := string(data), test.ExpectXML; got != want {
			if strings.Contains(want, "\n") {
				t.Errorf("#%d: marshal(%#v):\nHAVE:\n%s\nWANT:\n%s", idx, test.Value, got, want)
			} else {
				t.Errorf("#%d: marshal(%#v):\nhave %#q\nwant %#q", idx, test.Value, got, want)
			}
		}
	}
}

var extraNamespacesTests = []struct {
	Value              interface{}
	ExpectXML          string
	Namespaces         map[string]string
	OptimizeNamespaces bool
	PrefixElements     bool
}{
	// Test predefining the XMLSchema-Instance URL
	{
		Value: RequiredFieldStruct{},
		Namespaces: map[string]string{
			"xsi": "http://www.w3.org/2001/XMLSchema-instance",
		},
		ExpectXML: `<RequiredFieldStruct xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><SomeStruct xsi:nil="true" /></RequiredFieldStruct>`,
	},
	// Test namespaces without optimization
	{
		Value: struct {
			XMLName Name        `xml:"space Parent"`
			Child   *SomeStruct `xml:"space Child,!"`
		}{
			Child: &SomeStruct{},
		},
		ExpectXML: `<Parent xmlns="space"><Child xmlns="space"><Name></Name></Child></Parent>`,
	},
	// Test namespaces with optimization
	{
		Value: struct {
			XMLName Name        `xml:"space Parent"`
			Child   *SomeStruct `xml:"space Child,!"`
		}{
			Child: &SomeStruct{},
		},
		OptimizeNamespaces: true,
		ExpectXML:          `<Parent xmlns="space"><Child><Name></Name></Child></Parent>`,
	},
	// Test use prefixes
	{
		Value:          &SomeStructWithNSFields{},
		PrefixElements: true,
		ExpectXML:      `<SomeStructWithNSFields><Name xmlns="space"></Name></SomeStructWithNSFields>`,
	},
	// Test use prefixes with predefined namespace
	{
		Value:              &SomeStructWithNSFields{},
		OptimizeNamespaces: true,
		PrefixElements:     true,
		Namespaces: map[string]string{
			"s": "space",
		},
		ExpectXML: `<SomeStructWithNSFields xmlns:s="space"><s:Name></s:Name></SomeStructWithNSFields>`,
	},
}

func TestNamespaceExtras(t *testing.T) {
	for idx, test := range extraNamespacesTests {
		w := bytes.NewBuffer(nil)
		e := NewEncoder(w)
		e.OptimizeNamespaces(test.OptimizeNamespaces)
		e.PrefixElements(test.PrefixElements)
		for prefix, url := range test.Namespaces {
			e.Namespace(prefix, url)
		}

		err := e.Encode(test.Value)
		if err != nil {
			t.Errorf("#%d: marshal(%#v): %s", idx, test.Value, err)
			continue
		}
		_ = e.Flush()
		data := w.Bytes()

		if got, want := string(data), test.ExpectXML; got != want {
			if strings.Contains(want, "\n") {
				t.Errorf("#%d: marshal(%#v):\nHAVE:\n%s\nWANT:\n%s", idx, test.Value, got, want)
			} else {
				t.Errorf("#%d: marshal(%#v):\nhave %#q\nwant %#q", idx, test.Value, got, want)
			}
		}
	}
}

func TestCharDataReader(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testDataReader := bytes.NewBufferString("<Message>Hello world!</Message>")
	d := NewDecoder(testDataReader)

	tok, err := d.Token() // start element
	require.NoError(err)
	assert.Equal(StartElement{Name: Name{Local: "Message"}, Attr: []Attr{}}, tok)

	reader := d.CharDataReader()
	data, err := ioutil.ReadAll(reader)
	assert.Equal("Hello world!", string(data))

	tok, err = d.Token() // end element
	require.NoError(err)
	assert.Equal(EndElement{Name: Name{Local: "Message"}}, tok)
}
