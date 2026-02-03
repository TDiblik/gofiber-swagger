package gofiberswagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceNthOccurrence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		s        string
		old      string
		new      string
		n        int
		expected string
	}{
		{"a.b.c.d", ".", "-", 2, "a.b-c.d"},
		{"a.b.c.d", ".", "-", 1, "a-b.c.d"},
		{"a.b.c.d", ".", "-", 3, "a.b.c-d"},
		{"a.b.c.d", ".", "-", 4, "a.b.c.d"},
		{"a.b.c.d", ".", "-", 0, "a.b.c.d"},
		{"a.b.c.d", "x", "-", 1, "a.b.c.d"},
		{"", ".", "-", 1, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()
			result := replaceNthOccurrence(tc.s, tc.old, tc.new, tc.n)
			assert.Equal(t, tc.expected, result)
		})
	}
}
