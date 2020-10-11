package cli

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/env0/terratag/errors"
)

type Args struct {
	Tags                string
	Dir                 string
	SkipTerratagFiles   string
	IsSkipTerratagFiles bool
	Verbose             bool
}

func InitArgs() (Args, bool) {
	args := Args{}
	isMissingArg := false

	args.Tags = setFlag("tags", "")
	args.Dir = setFlag("dir", ".")
	skipTerratagFiles := setFlag("skipTerratagFiles", "true")
	verbose := setFlag("verbose", "false")

	if args.Tags == "" {
		log.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

	var err error
	args.IsSkipTerratagFiles, err = strconv.ParseBool(skipTerratagFiles)
	errorMessage := "-skipTerratagFiles may only be set to true or false"
	errors.PanicOnError(err, &errorMessage)

	args.Verbose, err = strconv.ParseBool(verbose)
	errorMessage2 := "-verbose may only be set to true or false"
	errors.PanicOnError(err, &errorMessage2)

	return args, isMissingArg
}

func setFlag(flag string, defaultValue string) string {
	result := defaultValue
	prefix := "-" + flag + "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			result = strings.TrimPrefix(arg, prefix)
		}
	}

	return result
}
