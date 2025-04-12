package main

import (
	"fmt"
	"os"

	"github.com/carlosrabelo/pbasic/pbasic/internal/repl"
)

func main() {
	r := repl.New(os.Stdin, os.Stdout)
	r.Run()
	fmt.Println()
}
