package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/env0/terratag"
	"github.com/env0/terratag/cli"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"
)

var version = "dev"

func main() {
	args, err := cli.InitArgs()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")

		os.Exit(2)
	}

	if args.Version {
		var versionPrefix string

		if !strings.HasPrefix(version, "v") {
			versionPrefix = "v"
		}

		fmt.Printf("Terratag %s%s\n", versionPrefix, version)

		return
	}

	initLogFiltering(args.Verbose)

	if err := terratag.Terratag(args); err != nil {
		log.Printf("[ERROR] execution failed due to an error\n%v", err)
		os.Exit(1)
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
