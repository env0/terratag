package terratag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHclMap(t *testing.T) {
	validCases := map[string]string{
		`{"a":"b","c":"d"}`: `{"a"="b","c"="d"}`,
		`a=b,c=d`:           `{"a"="b","c"="d"}`,
		`a-key=b-value`:     `{"a-key"="b-value"}`,
		"{}":                "{}",
	}

	for input, output := range validCases {
		input, expectedOutput := input, output
		t.Run("valid input "+input, func(t *testing.T) {
			output, err := toHclMap(input)
			assert.NoError(t, err)
			assert.Equal(t, expectedOutput, output)
		})
	}

	invalidCases := []string{
		"a$#$=b",
		`{"a": {"b": "c"}}`,
		"_a=b",
		"5a=b",
		"a=b!",
	}

	for i := range invalidCases {
		input := invalidCases[i]
		t.Run("invalid input "+input, func(t *testing.T) {
			_, err := toHclMap(input)
			assert.Error(t, err)
		})
	}
}
