package main

import (
	"fmt"
	"os"
	"site-mirror/internal/downloader"
	"site-mirror/internal/parser"
)

func outErrAndExit(err error) {
	_, err = fmt.Fprintln(os.Stderr, err)
	if err != nil {
		panic(err)
	}
	os.Exit(1)
}

func main() {
	cfg, err := parser.ParseArgs()
	if err != nil {
		outErrAndExit(err)
	}

	dwnld := downloader.NewDownloader()
	body, ctype, err := dwnld.Download(cfg.StartURL)
	if err != nil {
		outErrAndExit(err)
	}

	fmt.Printf("Content-Type: %s\n", ctype)
	fmt.Printf("Content-Length: %d\n", len(body))

	f, err := os.Create("index.html")
	if err != nil {
		outErrAndExit(err)
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {

		}
	}(f)
	_, err = f.Write(body)
	if err != nil {
		outErrAndExit(err)
	}
}
