package main

import (
	"fmt"
	"os"

	"github.com/carlosrabelo/pbasic/pbasic/internal/repl"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: pbasic [file.bas]")
		os.Exit(1)
	}

	r := repl.New(os.Stdin, os.Stdout)
	if len(os.Args) == 1 {
		r.Run()
		fmt.Println()
		return
	}

	if err := r.RunFile(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
