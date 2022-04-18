package convert

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

var localsRegex = regexp.MustCompile(`"([^"]*)"[ ]*=[ ]*"([^"]*)"`)

type Locals map[string]string

func decodeLocals(locals Locals, s string) error {
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

func encodeLocals(locals Locals) string {
	ret := "{"
	for key, value := range locals {
		ret += fmt.Sprintf("\"%s\" = \"%s\", ", key, value)
	}

	ret = strings.TrimSuffix(ret, ", ")
	ret += "}"
	return ret
}

func MergeLocals(attribute *hclwrite.Attribute, added string) (string, error) {
	localsAttribute := Locals{}
	tokens := hclwrite.Tokens{}
	existingLocalsExpression := stringifyExpression(attribute.BuildTokens(tokens))
	if err := decodeLocals(localsAttribute, existingLocalsExpression); err != nil {
		return "", err
	}

	localsAdded := Locals{}
	if err := decodeLocals(localsAdded, added); err != nil {
		return "", err
	}

	for key, value := range localsAdded {
		localsAttribute[key] = value
	}

	return encodeLocals(localsAttribute), nil
}
