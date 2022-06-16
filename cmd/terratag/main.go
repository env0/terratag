package main

import (
	"log"
	"os"

	"github.com/env0/terratag"
	"github.com/env0/terratag/cli"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"
)

func main() {
	args, err := cli.InitArgs()
	if err != nil {
		log.Println(err)
		log.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")
		return
	}
	initLogFiltering(args.Verbose)

	if err := terratag.Terratag(args); err != nil {
		log.Printf("[ERROR] execution failed due to an error\n%v", err)
	}
}

func initLogFiltering(verbose bool) {
	level := "INFO"
	if verbose {
		level = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	hclog.DefaultOutput = filter
}
