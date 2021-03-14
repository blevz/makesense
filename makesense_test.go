package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestJsonOutput(t *testing.T) {
	type JsonOutputTest struct {
		Input          string
		ExpectedTarget map[string]interface{}
	}
	tests := []JsonOutputTest{
		{
			Input: "testdata/c",
			ExpectedTarget: map[string]interface{}{
				"Targets": map[string]interface{}{
					"<ROOT>": map[string]interface{}{
						"MustRemake": false,
						"Name":       "<ROOT>",
					},
					"Makefile": map[string]interface{}{
						"MustRemake": false,
						"Name":       "Makefile",
					},
					"hellofunc.c": map[string]interface{}{
						"MustRemake": false,
						"Name":       "hellofunc.c",
					},
					"hellofunc.o": map[string]interface{}{
						"Cmds":       []interface{}{"gcc -c -o hellofunc.o hellofunc.c -I."},
						"MustRemake": true,
						"Name":       "hellofunc.o",
					},
					"hellomake": map[string]interface{}{
						"Cmds":       []interface{}{"gcc -o hellomake hellomake.o hellofunc.o -I."},
						"MustRemake": true,
						"Name":       "hellomake",
					},
					"hellomake.c": map[string]interface{}{
						"MustRemake": false,
						"Name":       "hellomake.c",
					},
					"hellomake.h": map[string]interface{}{
						"MustRemake": false,
						"Name":       "hellomake.h",
					},
					"hellomake.o": map[string]interface{}{
						"Cmds":       []interface{}{"gcc -c -o hellomake.o hellomake.c -I."},
						"MustRemake": true,
						"Name":       "hellomake.o",
					},
				},
			},
		},
		{
			Input: "testdata/basic",
			ExpectedTarget: map[string]interface{}{
				"Targets": map[string]interface{}{
					"<ROOT>": map[string]interface{}{
						"MustRemake": false,
						"Name":       "<ROOT>",
					},
					"Makefile": map[string]interface{}{
						"MustRemake": false,
						"Name":       "Makefile",
					},
					"a": map[string]interface{}{
						"Cmds":       []interface{}{"echo a"},
						"MustRemake": true,
						"Name":       "a",
					},
					"b": map[string]interface{}{
						"Cmds":       []interface{}{"echo b"},
						"MustRemake": true,
						"Name":       "b",
					},
					"c": map[string]interface{}{
						"Cmds":       []interface{}{"echo c", "echo multi", "echo line"},
						"MustRemake": true,
						"Name":       "c",
					},
					"test.txt": map[string]interface{}{
						"Cmds":       []interface{}{"# Testing comments", "echo \"hello\" >test.txt "},
						"MustRemake": true,
						"Name":       "test.txt",
					},
				},
			},
		},
	}
	for _, test := range tests {
		output, err := exec.Command("make", "-C", test.Input, "-Bnd").CombinedOutput()
		if err != nil {
			t.Error(string(output))
			t.Fatal(err)
		}
		g := &MakesenseGraph{
			Targets: map[string]*target{},
		}
		scanner := bufio.NewScanner(bytes.NewReader(output))
		root := g.GetTarget("<ROOT>")
		g.GraphScan(root, scanner, 0)
		var b strings.Builder
		g.dump(JSON, &b)
		var m map[string]interface{}
		json.NewDecoder(strings.NewReader(b.String())).Decode(&m)
		if diff := cmp.Diff(m, test.ExpectedTarget); diff != "" {
			t.Errorf("Diff: %s", diff)
		}
	}
}
