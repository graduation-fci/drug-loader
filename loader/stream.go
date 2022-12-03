package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/graduation-fci/phase1-demo/model"
)

type Element struct {
	Error error
	Drug  model.Drug
}

type jsonStream struct {
	stream chan Element
	path   string
}

func NewJsonStream(path string) *jsonStream {
	return &jsonStream{
		stream: make(chan Element),
		path:   path,
	}
}

func (j jsonStream) Watch() <-chan Element {
	return j.stream
}

func (j jsonStream) Pipeline() {
	defer close(j.stream)
	jsonFile, err := os.Open(j.path)
	if err != nil {
		j.stream <- Element{Error: fmt.Errorf("can't open file: %w", err)}
		return
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	decoder.Token() // Reading Open delimiter.

	i := 1
	for decoder.More() {
		var tempDrug model.Drug
		if err := decoder.Decode(&tempDrug); err != nil {
			j.stream <- Element{Error: fmt.Errorf("can't decode line %d: %w", i, err)}
			return
		}
		j.stream <- Element{Drug: tempDrug}
		i++
	}
	decoder.Token() // Read closing delimiter.
}
