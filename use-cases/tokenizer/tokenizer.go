package tokenizer

import (
	"fmt"
	"strings"
)

type ITokenizer interface {
	Handle() ([]string, error)
}

type Tokenizer struct {
	text string
	chunk int
	overlap int
}

func NewTokenizer(text string, chunk, overlap int) ITokenizer {
	return &Tokenizer{
		text: text,
		chunk: chunk,
		overlap: overlap,
	}
}

func (t *Tokenizer) Handle() ([]string, error) {
	fmt.Println("Tokenizer : Starting ...")

	tokens := strings.Split(t.text, " ")
	// tokens := []string{"A", "B", "C", "D", "E", "F"}

	output := []string{}

	count := 0
	str := ""
	for i := 0; i < len(tokens); i++ {
		str = str + " " + tokens[i]
		count++

		if t.chunk < count {
			output = append(output, str)
			str = ""
			count = 0
			i = i - t.overlap
		}
	}
	
	fmt.Println("Tokenizer : Completed")
	return output, nil
}
