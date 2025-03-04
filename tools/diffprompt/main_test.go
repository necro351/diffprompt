package main

import (
	"testing"
)

func TestDiff(t *testing.T) {
	var (
		a = `Hello, world!
It is a beautiful day.
`
		b = `Hello, world!
It is a beautiful night.
`
		expected = ` Hello, world!
-It is a beautiful day.
+It is a beautiful night.
`
	)

	diff, err := sideBySideDiff(a, b)
	if err != nil {
		t.Fatalf("sideBySideDiff failed: %v", err)
	}

	if diff != expected {
		t.Fatalf("diff = %q, expected %q", diff, expected)
	}
}
