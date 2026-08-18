package main

import (
	"fmt"
	"os"

	"github.com/jteutenberg/knob/session"
)

func main() {
	s := session.New(os.Stdout, os.Stderr)
	if err := s.Run(os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
