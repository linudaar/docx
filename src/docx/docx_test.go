package docx_test

import (
	"bufio"
	"docx"
	"io/ioutil"
	"os"
	"testing"
	"time"
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

	docx1.Replace("AgendaHeader", "My Cool Agenda", 1)
	docx1.Replace("MeetingDate", "1.1.2017", 1)
	docx1.ReplaceLoop("topic", topics)

	f, err := os.Create("output-" + time.Now().Format("20060102150405") + ".docx")
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)
	docx1.Write(w)

}
