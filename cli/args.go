package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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
	DefaultToTerraform  bool
	KeepExistingTags    bool
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
	fs.StringVar(&args.Type, "type", string(common.Terraform), "The IAC type. Valid values: terraform, terragrunt, or terragrunt-run-all")
	fs.BoolVar(&args.Version, "version", false, "Prints the version")
	fs.BoolVar(&args.DefaultToTerraform, "default-to-terraform", false, "By default uses OpenTofu (if installed), if set will use Terraform even when Opentofu is installed")
	fs.BoolVar(&args.KeepExistingTags, "keep-existing-tags", false, "When set, existing tags will be preserved when merging tags (by default, new tags override existing ones)")

	// Set cli args based on environment variables.
	// The command line flags have precedence over environment variables.
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "version" {
			return
		}

		name := "TERRATAG_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if value, ok := os.LookupEnv(name); ok {
			if err := fs.Set(f.Name, value); err != nil {
				fmt.Printf("[WARN] failed to set command arg flag '%s' from environment variable '%s': %v\n", f.Name, name, err)
			}
		}
	})

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
