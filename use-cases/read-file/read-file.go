package readfile

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

type IReadFile interface {
	Handle() (string, error)
}

type ReadFile struct {
	filePath string
}

func NewReadFile(filePath string) IReadFile {
	return &ReadFile{
		filePath: filePath,
	}
}

func (r *ReadFile) Handle() (string, error) {
	fmt.Println("ReadFile : Starting read for", r.filePath)

	if strings.HasSuffix(strings.ToLower(r.filePath), ".pdf") {
		f, pdfReader, err := pdf.Open(r.filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open PDF file: %w", err)
		}
		defer f.Close()

		var buf bytes.Buffer
		b, err := pdfReader.GetPlainText()
		if err != nil {
			return "", fmt.Errorf("failed to get plain text from PDF: %w", err)
		}
		_, err = buf.ReadFrom(b)
		if err != nil {
			return "", fmt.Errorf("failed to read text buffer: %w", err)
		}
		fmt.Println("ReadFile : Completed reading PDF successfully")
		return buf.String(), nil
	}

	// Fallback for plain text files
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}

	fmt.Println("ReadFile : Completed reading text file successfully")
	return string(data), nil
}