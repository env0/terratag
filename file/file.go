package file

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/env0/terratag/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func ReplaceWithTerratagFile(path string, textContent string, rename bool) {
	backupFilename := path + ".bak"

	if rename {
		taggedFilename := strings.TrimSuffix(path, filepath.Ext(path)) + ".terratag.tf"
		log.Print("[INFO] Creating file ", taggedFilename)
		taggedFileError := ioutil.WriteFile(taggedFilename, []byte(textContent), 0644)
		errors.PanicOnError(taggedFileError, nil)
	}

	log.Print("[INFO] Renaming original file from ", path, " to ", backupFilename)
	backupFileError := os.Rename(path, backupFilename)
	errors.PanicOnError(backupFileError, nil)

	if !rename {
		log.Print("[INFO] Creating file ", path)
		taggedFileError := ioutil.WriteFile(path, []byte(textContent), 0644)
		errors.PanicOnError(taggedFileError, nil)
	}
}

func GetFilename(path string) string {
	_, filename := filepath.Split(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(path))
	filename = strings.ReplaceAll(filename, ".", "-")
	return filename
}

func ReadHCLFile(path string) *hclwrite.File {
	src, err := ioutil.ReadFile(path)
	errors.PanicOnError(err, nil)

	file, diagnostics := hclwrite.ParseConfig(src, path, hcl.InitialPos)
	if diagnostics.HasErrors() {
		hclErrors := diagnostics.Errs()
		log.Fatalln(hclErrors)
	}
	return file
}
