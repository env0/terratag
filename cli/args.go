package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/env0/terratag/internal/common"
)

type Args struct {
	Tags                string
	Dir                 string
	Filter              string
	Skip                string
	Type                string
	IsSkipTerratagFiles bool
	Verbose             bool
	Rename              bool
	Version             bool
}

func validate(args Args) error {
	if args.Tags == "" {
		return errors.New("missing tags")
	}
	if args.Type != string(common.Terraform) && args.Type != string(common.Terragrunt) {
		return fmt.Errorf("invalid type %s, must be either 'terratag' or 'terragrunt'", args.Type)
	}
	return nil
}

func InitArgs() (Args, error) {
	args := Args{}
	programName := os.Args[0]
	programArgs := os.Args[1:]

	fs := flag.NewFlagSet(programName, flag.ExitOnError)

	fs.StringVar(&args.Tags, "tags", "", "Tags as a valid JSON document")
	fs.StringVar(&args.Dir, "dir", ".", "Directory to recursively search for .tf files and terratag them")
	fs.BoolVar(&args.IsSkipTerratagFiles, "skipTerratagFiles", true, "Skips any previously tagged files")
	fs.StringVar(&args.Filter, "filter", ".*", "Only apply tags to the selected resource types (regex)")
	fs.StringVar(&args.Skip, "skip", "", "Exclude the selected resource types from tagging (regex)")
	fs.BoolVar(&args.Verbose, "verbose", false, "Enable verbose logging")
	fs.BoolVar(&args.Rename, "rename", true, "Keep the original filename or replace it with <basename>.terratag.tf")
	fs.StringVar(&args.Type, "type", string(common.Terraform), "The IAC type. Valid values: terraform or terragrunt")
	fs.BoolVar(&args.Version, "version", false, "Prints the version")

	if err := fs.Parse(programArgs); err != nil {
		return args, err
	}

	if args.Version {
		return args, nil
	}

	if err := validate(args); err != nil {
		return args, err
	}

	return args, nil
}
