package main

import (
	"log"
	"os"

	"github.com/env0/terratag/cli"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"
)

func main() {
	args, isMissingArg := cli.InitArgs()
	if isMissingArg {
		return
	}
	initLogFiltering(args.Verbose)

	cli.Terratag(args)
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
