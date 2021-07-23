package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKey(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "Key can end on number",
			key:   "TheAnswerIs42",
			valid: true,
		},
		{
			name:  "Key can container underscores and dashes",
			key:   "This-key_is-valid",
			valid: true,
		},
		{
			name:  "Key cannot be empty",
			key:   "",
			valid: false,
		},
		{
			name:  "Key contains an invalid character",
			key:   "foo!",
			valid: false,
		},
		{
			name:  "First character of key is not a letter",
			key:   "2BeOrNot2Be",
			valid: false,
		},
		{
			name:  "Last character of key is not a letter of number",
			key:   "invalid-",
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateKey(tc.key)

			if tc.valid {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.IsType(t, &InvalidKey{}, err)
				}
			}
		})
	}
}
