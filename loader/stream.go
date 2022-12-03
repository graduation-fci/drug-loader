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
}

func NewJsonStream() *jsonStream {
	return &jsonStream{
		stream: make(chan Element),
	}
}

func (j jsonStream) Watch() <-chan Element {
	return j.stream
}

func (j jsonStream) Start(path string) {
	defer close(j.stream)
	jsonFile, err := os.Open(path)
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
