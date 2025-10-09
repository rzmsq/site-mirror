package main

import (
	"fmt"
	"os"
	"site-mirror/internal/downloader"
	"site-mirror/internal/parser"
	"site-mirror/internal/storage"
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

	st := storage.NewStorage(cfg.OutputDir)
	err = st.Save(cfg.StartURL, body, ctype)
	if err != nil {
		outErrAndExit(err)
	}

	pars := parser.NewParser()
	pages, links, err := pars.ParseHTML(body, cfg.StartURL)
	if err != nil {
		outErrAndExit(err)
	}
	fmt.Printf("Found %d pages: %v\n", len(pages), pages)
	fmt.Printf("Found %d links: %v\n", len(links), links)
}
