package readfile

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/gen2brain/go-fitz"
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
		doc, err := fitz.New(r.filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open PDF file: %w", err)
		}
		defer doc.Close()

		var buf bytes.Buffer
		numPages := doc.NumPage()
		fmt.Printf("ReadFile : Parsing %d pages from PDF...\n", numPages)

		for pageNum := 0; pageNum < numPages; pageNum++ {
			text, err := doc.Text(pageNum)
			if err != nil {
				return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum+1, err)
			}
			buf.WriteString(text)
			buf.WriteString("\n")
		}

		fmt.Println("ReadFile : Completed reading PDF successfully using go-fitz")
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