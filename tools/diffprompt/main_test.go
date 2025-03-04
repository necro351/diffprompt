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

func TestApply(t *testing.T) {
	var (
		input = `apply
+line1
-line2
 line3
revert
+line4
-line5
 line6`
		expected = `line1
line3
revert
line4
line6`
	)

	result := applyCommands(input)
	if result != expected {
		t.Error(result, expected)
	}
}

func TestApplyNone(t *testing.T) {
	var (
		input = `+line1
-line2
 line3
revert
+line4
-line5
 line6`
		expected = `+line1
-line2
 line3
revert
+line4
-line5
 line6`
	)

	result := applyCommands(input)
	if result != expected {
		t.Error(result, expected)
	}
}

func TestApplyReject(t *testing.T) {
	var (
		input = `apply
+line1
-line2
 line3
           reject         
+line4
-line5
 line6`
		expected = `line1
line3
line5
line6`
	)

	result := applyCommands(input)
	if result != expected {
		t.Error(result, expected)
	}
}

func TestApplyEmpty(t *testing.T) {
	var (
		input    = ``
		expected = `
`
	)

	result := applyCommands(input)
	if result != expected {
		t.Error(result, expected)
	}
}

func TestApplyEmptyBlock(t *testing.T) {
	var (
		input = `apply

		reject`
		expected = `
`
	)

	result := applyCommands(input)

	if result != expected {
		t.Error(result, expected)
	}
}
