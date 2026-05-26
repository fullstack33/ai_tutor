package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type InputHandler struct {
	reader *bufio.Reader
}

func NewInputHandler() *InputHandler {
	return &InputHandler{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (h *InputHandler) Ask(prompt string) (string, error) {
	fmt.Print(prompt)
	text, err := h.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
