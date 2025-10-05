package main

import (
	"fmt"
	"os"
	"site-mirror/internal/parser"
)

func main() {
	cfg, err := parser.ParseArgs()
	if err != nil {
		_, err = fmt.Fprintln(os.Stderr, err)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	fmt.Println(cfg)
}
