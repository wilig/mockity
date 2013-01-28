package main

import (
	"strings"
	"testing"
)

func TestPreProcessConcatsMultiLineStringLiterals(t *testing.T) {
	src := "{\"body\": \"this is\na multiline\nstring\"}"
	out := preProcess([]byte(src))
	if strings.IndexRune(string(out), '\n') != -1 {
		t.Error("Failed to rewrite newline to space.")
	}
}

func TestPreProcessHandlesPartialLineComments(t *testing.T) {
	src := "{\"body\": \"value\"} // what a meaningful body!"
	out := preProcess([]byte(src))
	if strings.Index(string(out), "// what") != -1 {
		t.Error("Failed to drop comment from end of buffer.")
	}
	src = "{\"body\": // value comes next\n\"value\"}"
	out = preProcess([]byte(src))
	if strings.Index(string(out), "// value") != -1 {
		t.Error("Failed to drop comment from middle of buffer.")
	}
}

func TestPreProcessHandlesFullLineComments(t *testing.T) {
	src := "// My config\n{\"body\": \"value\"}"
	out := preProcess([]byte(src))
	if strings.Index(string(out), "// My config") != -1 {
		t.Error("Failed to drop full line comment.")
	}
}

func TestPreProcessIgnoresHashesInStringLiterals(t *testing.T) {
	src := "# My config\n{\"body\": \"value //look //find\"}"
	out := preProcess([]byte(src))
	if strings.Index(string(out), "//look") == -1 {
		t.Error("Failed to ignore comment inside string literal.")
	}
	src = "# My config\n{\"body\": \"value \n//look \n//find\"}"
	out = preProcess([]byte(src))
	if strings.Index(string(out), "//") == -1 {
		t.Error("Failed to ignore comment inside string literal on new line.")
	}
}
