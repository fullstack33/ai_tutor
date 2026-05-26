package main

import (
	"fmt"
	"github.com/fullstack33/ai-tutor/use-cases/orachastrator"
)

func main() {
	fmt.Println("Started ...")

	orachastrator.NewOrachastrator().Handle()

	fmt.Println("Completed ...")
}
