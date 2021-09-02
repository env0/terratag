package cli

import (
	"fmt"
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
	Filter              []string
	IsSkipTerratagFiles bool
	Verbose             bool
	Rename              bool
}

func InitArgs() (Args, bool) {
	args := Args{}
	isMissingArg := false

	args.Tags = setFlag("tags", "")
	args.Dir = setFlag("dir", ".")
	args.IsSkipTerratagFiles = booleanFlag("skipTerratagFiles", true)
	args.Filter = setArrayFlag("filter", make([]string, 0))
	args.Verbose = booleanFlag("verbose", false)
	args.Rename = booleanFlag("rename", true)

	if args.Tags == "" {
		log.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

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

func setArrayFlag(flag string, defaultValue []string) []string {
	result := defaultValue
	prefix := "-" + flag + "="
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			result = strings.Split(strings.TrimPrefix(arg, prefix), ",")
		}
	}

	return result
}

func booleanFlag(flag string, defaultValue bool) bool {
	defaultString := "false"
	if defaultValue {
		defaultString = "true"
	}
	stringValue := setFlag(flag, defaultString)
	value, err := strconv.ParseBool(stringValue)
	errorMessage := fmt.Sprint("-", flag, " may only be set to true or false")
	errors.PanicOnError(err, &errorMessage)
	return value
}
