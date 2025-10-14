package main

import (
	"errors"
	"fmt"
	"os"
	"site-mirror/internal/config"
	"site-mirror/internal/downloader"
	"site-mirror/internal/parser"
	"site-mirror/internal/queue"
	"site-mirror/internal/storage"
	"sync"
)

func printErrAndExit(err error) {
	_, err = fmt.Fprintln(os.Stderr, err)
	if err != nil {
		panic(err)
	}
	os.Exit(1)
}

func main() {
	if err := runApp("SiteMirror"); err != nil {
		printErrAndExit(err)
	}
}

func runApp(userAgent string) error {
	cfg, err := parser.ParseArgs()
	if err != nil {
		return err
	}

	q := queue.NewQueue(1000, cfg.StartURL.Host)
	dwnld, err := downloader.NewDownloader(cfg.StartURL, userAgent)
	if err != nil {
		return err
	}
	st := storage.NewStorage(cfg.OutputDir)
	pars := parser.NewParser()

	wg := &sync.WaitGroup{}
	wg.Add(cfg.Concurrency)
	for range cfg.Concurrency {
		go runWorker(q, pars, dwnld, st, cfg, wg)
	}

	initTask := queue.Task{URL: cfg.StartURL, Depth: 0, Type: "page"}
	err = q.Enqueue(initTask, cfg.Depth)
	if err != nil {
		return err
	}

	fmt.Println("Processing...")
	q.WaitAndClose()
	fmt.Println("Done")
	return nil
}

func runWorker(q *queue.Queue, pars *parser.Parser, dwnld *downloader.Downloader, st *storage.Storage, cfg *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range q.Dequeue() {
		body, ctype, err := dwnld.Download(task.URL, cfg.UseRobots)
		if err != nil && !errors.Is(err, downloader.ErrTooManyAttempts) {
			printErrAndExit(err)
		}

		err = st.Save(task.URL, body, ctype)
		if err != nil {
			printErrAndExit(err)
		}

		if task.Depth < cfg.Depth {
			pages, resources, errParser := pars.ParseHTML(body, task.URL)
			if errParser != nil {
				printErrAndExit(errParser)
			}
			for _, page := range pages {
				newTask := queue.Task{URL: page, Depth: task.Depth + 1, Type: "page"}
				err = q.Enqueue(newTask, cfg.Depth)
				if err != nil {
					continue
				}
			}
			for _, resource := range resources {
				newTask := queue.Task{URL: resource, Depth: task.Depth + 1, Type: "page"}
				err = q.Enqueue(newTask, cfg.Depth)
				if err != nil {
					continue
				}
			}
		}
		q.Done()
	}
}
