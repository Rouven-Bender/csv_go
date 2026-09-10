package csv_go_test

import (
	"os"
	"testing"

	"github.com/Rouven-Bender/csv_go"
)

type Content struct {
	Betrag float64 `csv:"A"`
	Text string `csv:"b"`
	Enum int `csv:"c"`
}

func TestReaderRead(t *testing.T) {
	f, err := os.Open("./testdata/eins.csv")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	r := csv_go.NewReader[Content](f)
	r.SetComma(';')
	r.SetGermanDecimal(true)
	err = r.SkipLine() // Skip header with column names
	if err != nil {
		t.Error(err)
	}
	c, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if c.Text != "Hallo" {
		t.Log(c)
		t.Error("Wrong text")
	}
}

func TestReaderIter(t *testing.T) {
	f, err := os.Open("./testdata/eins.csv")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	r := csv_go.NewReader[Content](f)
	r.SetComma(';')
	r.SetGermanDecimal(true)
	err = r.SkipLine() // Skip header with column names
	if err != nil {
		t.Error(err)
	}
	
	sum := 0
	for e := range r.Iter() {
		sum += e.Enum
	}
	if sum != 50 {
		t.Errorf("wrong sum got: %d", sum)
	}
}

//func TestWriterWrite(t *testing.T) {
//	f, err := os.Create("/dev/shm/gotest.csv")
//	if err != nil {
//		t.Error(err)
//	}
//	w := csv_go.NewWriter[Content](f)
//	defer w.Close()
//	w.SetComma(';')
//	c := Content{Betrag: 1.30, Text: "Works", Enum: 69420}
//	err = w.Write(&c)
//	if err != nil {
//		t.Error(err)
//	}
//	w.Flush()
//}
