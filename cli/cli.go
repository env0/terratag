package cli

import (
	"github.com/env0/terratag/errors"
	"log"
	"os"
	"strconv"
	"strings"
)

func InitArgs() (string, string, bool, bool) {
	var tags string
	var dir string
	var skipTerratagFiles string
	var isSkipTerratagFiles bool

	isMissingArg := false

	tags = setFlag("tags", "")
	dir = setFlag("dir", ".")
	skipTerratagFiles = setFlag("skipTerratagFiles", "true")

	if tags == "" {
		log.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")
		isMissingArg = true
	}

	isSkipTerratagFiles, err := strconv.ParseBool(skipTerratagFiles)
	errorMessage := "-skipTerratagFiles may only be set to true or false"
	errors.PanicOnError(err, &errorMessage)

	return tags, dir, isSkipTerratagFiles, isMissingArg
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
