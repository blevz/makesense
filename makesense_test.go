package main

import "testing"

func TestIndentAndTrim(t *testing.T) {
	type IndentAndTrimTest struct {
		Input                 string
		ExpectedIndentation   int
		ExpectedTrimmedString string
	}
	tests := []IndentAndTrimTest{
		{
			Input:                 "   a",
			ExpectedIndentation:   3,
			ExpectedTrimmedString: "a",
		},
		{
			Input:                 "a b c",
			ExpectedIndentation:   0,
			ExpectedTrimmedString: "a b c",
		},
	}
	for _, test := range tests {
		trimmed, indentLevel := findIndentAndTrim(test.Input)
		if indentLevel != test.ExpectedIndentation {
			t.Errorf("Indent level should be %d was %d", test.ExpectedIndentation, indentLevel)
		}
		if trimmed != test.ExpectedTrimmedString {
			t.Errorf("Trimmed should be `%s` was `%s`", test.ExpectedTrimmedString, trimmed)
		}
	}
}

func TestTargetNameFromLine(t *testing.T) {
	type TargetNameFromLineTest struct {
		Input          string
		ExpectedTarget string
	}
	tests := []TargetNameFromLineTest{
		{
			Input:          "Considering target file `b'.",
			ExpectedTarget: "b",
		},
		{
			Input:          "Considering target file `my great target'.",
			ExpectedTarget: "my great target",
		},
		{
			Input:          "Considering target file 'my great target'.",
			ExpectedTarget: "my great target",
		},
	}
	for _, test := range tests {
		target := targetNameFromLine(test.Input)
		if target != test.ExpectedTarget {
			t.Errorf("Target should be `%s` was `%s`", test.ExpectedTarget, target)
		}
	}
}
