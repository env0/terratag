package cli

import (
	"flag"
	"log"
	"os"
)

type Args struct {
	Tags                string
	Dir                 string
	Filter              string
	IsSkipTerratagFiles bool
	Verbose             bool
	Rename              bool
	Terragrunt          bool
}

func InitArgs() (Args, bool) {
	args := Args{}
	isMissingArg := false
	programName := os.Args[0]
	programArgs := os.Args[1:]

	fs := flag.NewFlagSet(programName, flag.ExitOnError)

	fs.StringVar(&args.Tags, "tags", "", "Tags as a valid JSON document")
	fs.StringVar(&args.Dir, "dir", ".", "Directory to recursively search for .tf files and terratag them")
	fs.BoolVar(&args.IsSkipTerratagFiles, "skipTerratagFiles", true, "Skips any previously tagged files")
	fs.StringVar(&args.Filter, "filter", ".*", "Only apply tags to the selected resource types (regex)")
	fs.BoolVar(&args.Verbose, "verbose", false, "Enable verbose logging")
	fs.BoolVar(&args.Rename, "rename", true, "Keep the original filename or replace it with <basename>.terratag.tf")
	fs.BoolVar(&args.Terragrunt, "terragrunt", false, "Tags all the terraform files under .terragrunt-cache")

	err := fs.Parse(programArgs)

	if err != nil || args.Tags == "" {
		log.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

	return args, isMissingArg
}
