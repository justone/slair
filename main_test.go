package main

import (
	"testing"

	emoji "gopkg.in/kyokomi/emoji.v1"

	"github.com/stretchr/testify/assert"
)

func TestParseEmojis(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input  string
		output []string
		err    string
	}{
		{
			":boom:",
			[]string{emoji.Sprint(":boom:")},
			"",
		},
		{
			":invalid:",
			[]string{emoji.Sprint(":boom:")},
			"Unable to resolve emoji for :invalid:",
		},
		{
			":boom:,:tada:",
			[]string{
				emoji.Sprint(":boom:"),
				emoji.Sprint(":tada:"),
			},
			"",
		},
		{
			"  boom  , tada",
			[]string{
				emoji.Sprint(":boom:"),
				emoji.Sprint(":tada:"),
			},
			"",
		},
	}

	for _, test := range tests {
		emojis, err := ParseEmojis(test.input)

		if len(test.err) > 0 {
			assert.NotNil(err)
			assert.Contains(err.Error(), test.err)
		} else {
			assert.Nil(err)
			assert.Equal(test.output, emojis)
		}
	}
}
