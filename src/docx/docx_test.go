package docx_test

import (
	"docx"
	"io/ioutil"
	"testing"
)

func TestReplace(t *testing.T) {
	fileBytes, err := ioutil.ReadFile("template.docx")
	if err != nil {
		panic(err)
	}

	r, err := docx.ReadDoxFileFromBytes(fileBytes)
	if err != nil {
		panic(err)
	}

	docx1 := r.Editable()
	var topics = []map[string]string{
		map[string]string{
			"name": "topic A",
			"pos":  "TOP 01",
			"user": "Thomas Smith",
		},
		map[string]string{
			"name": "topic B",
			"pos":  "TOP 02",
			"user": "John Doe",
		},
	}

	docx1.ReplaceLoop("topic", topics)
	docx1.WriteToFile("output.docx")
}
