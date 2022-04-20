package convert

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

var localsRegex = regexp.MustCompile(`"([^"]*)"[ \t]*=[ \t]*"([^"]*)"`)

type Locals map[string]string

func decodeTerratagLocals(locals Locals, s string) error {
	for k := range locals {
		delete(locals, k)
	}

	matches := localsRegex.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return errors.New("no matches found when decoding locals")
	}

	for _, match := range matches {
		locals[match[1]] = match[2]
	}

	return nil
}

func encodeTerratagLocals(locals Locals) string {
	ret := "{"

	// Return it in ordered manner for order consistency (primarly when running tests).
	var keys []string
	for key := range locals {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ret += fmt.Sprintf("\"%s\" = \"%s\", ", key, locals[key])
	}

	ret = strings.TrimSuffix(ret, ", ")
	ret += "}"
	return ret
}

func MergeTerratagLocals(attribute *hclwrite.Attribute, added string) (string, error) {
	localsAttribute := Locals{}
	tokens := hclwrite.Tokens{}
	existingLocalsExpression := stringifyExpression(attribute.BuildTokens(tokens))
	if err := decodeTerratagLocals(localsAttribute, existingLocalsExpression); err != nil {
		return "", err
	}

	localsAdded := Locals{}
	if err := decodeTerratagLocals(localsAdded, added); err != nil {
		return "", err
	}

	for key, value := range localsAdded {
		localsAttribute[key] = value
	}

	return encodeTerratagLocals(localsAttribute), nil
}
