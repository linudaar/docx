package docx_test

import (
	"bufio"
	"docx"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var (
	RunDir string
)

func init() {
	RunDir = "output/run-" + time.Now().Format("20060102150405")
}

func TestMain(m *testing.M) {
	err := os.MkdirAll(RunDir, 0777)
	if err != nil {
		panic(err)
	}
	retCode := m.Run()
	os.Exit(retCode)
}

func TestReplace1(t *testing.T) {
	fmt.Println("TestReplace1: Open template.docx")
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

	var participants = []map[string]string{
		map[string]string{
			"name": "Hugo",
		},
		map[string]string{
			"name": "Franz",
		},
		map[string]string{
			"name": "Kurt",
		},
		map[string]string{
			"name": "Sepp",
		},
	}

	docx1.Replace("AgendaHeader", "My Cool Agenda", 1)
	docx1.Replace("MeetingDate", "1.1.2017", 1)
	docx1.ReplaceLoop("topic", topics)
	docx1.ReplaceLoop("participant", participants)

	f1, err := os.Create(RunDir + "/output-1.docx")
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f1)
	docx1.Write(w)

	f2, err := os.Create(RunDir + "/document-1.xml")
	if err != nil {
		panic(err)
	}
	f2.WriteString(docx1.Content)

}

func TestReplace2(t *testing.T) {
	fmt.Println("TestReplace1: Open output-1.docx")
	fileBytes, err := ioutil.ReadFile(RunDir + "/output-1.docx")
	if err != nil {
		panic(err)
	}

	r, err := docx.ReadDoxFileFromBytes(fileBytes)
	if err != nil {
		panic(err)
	}

	docx1 := r.Editable()

	var participants = []map[string]string{
		map[string]string{
			"name": "Hugo",
		},
		map[string]string{
			"name": "Franz",
		},
		map[string]string{
			"name": "Kurt",
		},
		map[string]string{
			"name": "Sepp",
		},
	}

	docx1.ReplaceLoop("participant", participants)
	f, err := os.Create(RunDir + "/output-2.docx")
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)
	docx1.Write(w)
	f2, err := os.Create(RunDir + "/document-2.xml")
	if err != nil {
		panic(err)
	}
	f2.WriteString(docx1.Content)
}
