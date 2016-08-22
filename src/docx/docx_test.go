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
			"name": "Grand Theory of Everything",
			"pos":  "TOP 01",
			"user": "Stephen Hawkings",
		},
		map[string]string{
			"name": "The Breakable Atom",
			"pos":  "TOP 02",
			"user": "Niels Bohr",
		},
		map[string]string{
			"name": "At the Speed of Light",
			"pos":  "TOP 03",
			"user": "Albert Einstein",
		},
		map[string]string{
			"name": "The Universe and the Rest",
			"pos":  "TOP 04",
			"user": "Isaac Newton",
		},
		map[string]string{
			"name": "Why Forty Two",
			"pos":  "TOP 05",
			"user": "Douglas Adams",
		},
	}

	var participants = []map[string]string{
		map[string]string{
			"name": "Albert Einstein",
		},
		map[string]string{
			"name": "Isaac Newton",
		},
		map[string]string{
			"name": "Niels Bohr",
		},
		map[string]string{
			"name": "Stephen Hawkings",
		},
		map[string]string{
			"name": "Douglas Adams",
		},
	}

	docx1.Replace("AgendaHeader", "On the Meaning of Life", 1)
	docx1.Replace("MeetingDate", "01.01.2017 (8:00)", 1)
	docx1.Replace("additionalInfo", "Dinner in the restaurant at the end of the galaxy!", 1)
	docx1.Replace("host", "Paranoid Android", 1)
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
