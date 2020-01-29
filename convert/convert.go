package convert

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thoas/go-funk"
	"strings"
)

func GetExistingTagsExpression(tokens hclwrite.Tokens) string {
	if isHclMap(tokens) {
		return buildMapExpression(tokens)
	} else {
		return stringifyExpression(tokens)
	}
}

func isHclMap(tokens hclwrite.Tokens) bool {
	maybeHclMap := strings.TrimSpace(string(tokens.Bytes()))
	return strings.HasPrefix(maybeHclMap, "{") && strings.HasSuffix(maybeHclMap, "}")
}

func buildMapExpression(tokens hclwrite.Tokens) string {
	// Need to convert to inline map expression

	// First, we remove the first and last two tokens - get rid of the openning { closing } and newline
	tokens = tokens[1:]
	tokens = tokens[:len(tokens)-2]

	// Then, we normalize the key-value paris so that they would be seperated by comma
	// That's cause HCL supports both newline, comma or a combination of the two in seperating values
	// This will make it easier to split the key-value pairs later
	var tokensToRemove hclwrite.Tokens
	for i, token := range tokens {
		if token.Type == hclsyntax.TokenNewline {
			// make sure there is (or was) no comma before
			if i > 0 && tokens[i-1].Type != hclsyntax.TokenComma && !funk.Contains(tokensToRemove, tokens[i-1]) {
				tokens[i] = &hclwrite.Token{
					Type:         hclsyntax.TokenComma,
					Bytes:        []byte(","),
					SpacesBefore: 1,
				}
			} else { // if there is, we should remove this new line, so we'll only have the comma
				tokensToRemove = append(tokensToRemove, token)
			}
		}
	}

	// Remove tall the new lines we marked for removal
	for _, tokenToRemove := range tokensToRemove {
		indexToRemove := funk.IndexOf(tokens, tokenToRemove)
		tokens = append(tokens[:indexToRemove], tokens[indexToRemove+1:]...)
	}

	// At this point there should be no new lines, only a single comma seperating between key values
	// Since the map() gets a flat set of pairs as map("key1","value1","key2","value2"),
	// we can just replace any assignment operator (=) with comma
	for i, token := range tokens {
		if token.Type == hclsyntax.TokenEqual {
			tokens[i] = &hclwrite.Token{
				Type:         hclsyntax.TokenComma,
				Bytes:        []byte(","),
				SpacesBefore: token.SpacesBefore,
			}
		}
	}

	mapContent := string(tokens.Bytes())
	mapContent = strings.TrimSpace(mapContent)
	mapContent = strings.TrimSuffix(mapContent, ",") // remove any traling commas due to newline replaced
	return "map(" + mapContent + ")"
}

func stringifyExpression(tokens hclwrite.Tokens) string {
	expression := strings.TrimSpace(string(tokens.Bytes()))
	// may be wrapped with ${ } in TF11
	expression = strings.TrimPrefix(expression, "${")
	expression = strings.TrimSuffix(expression, "${")

	return expression
}
