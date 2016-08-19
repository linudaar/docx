package docx_test

import (
	"docx"
	"testing"
)

func TestReplace(t *testing.T) {
	r, err := docx.ReadDocxFile("template.docx")
	if err != nil {
		panic(err)
	}

	docx1 := r.Editable()
	// docx1.Replace("old_1_1", "new_1_1", -1)
	// docx1.Replace("old_1_2", "new_1_2", -1)
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
