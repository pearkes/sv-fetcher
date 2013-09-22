package main

import (
	"testing"
)

func TestFetcher_TestContentNeeded(t *testing.T) {
	// Yes, we need it
	answer, ext := checkContentNeeded("foo.md")
	if answer != true {
		t.Fatal("answer should be true")
	}
	if ext != "md" {
		t.Fatal("extension should be md")
	}
	// Yes again
	answer, ext = checkContentNeeded("foo.txt")
	if answer != true {
		t.Fatal("answer should be true")
	}
	if ext != "txt" {
		t.Fatal("extension should be txt")
	}
	// No, we don't
	answer, ext = checkContentNeeded("foo.jpg")
	if answer != false {
		t.Fatal("answer should be false")
	}
	if ext != "" {
		t.Fatal("extension should be nil string")
	}
}

func TestFetcher_TestParseMetaData(t *testing.T) {
	file := "/project_01.jpg"
	filename, tag, order := parseMetaData(file)
	if filename != "project_01.jpg" {
		t.Fatalf("filename does not match: %s", filename)
	}
	if tag != "project" {
		t.Fatalf("tag does not match: %s", tag)
	}

	if order != 1 {
		t.Fatalf("order does not match: %s", order)
	}
}
