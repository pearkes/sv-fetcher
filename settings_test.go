package main

import (
	"errors"
	"testing"
)

var exampleFile = `# You can change these, and soon they will be read and updated
# Feel free to email broken@smallvictori.es if you need help :-)

domain: foo.example.com
title: Title One!

# Any errors will appear here

`

var badExampleFile = `# You can change these, and soon they will be read and updated
# Feel free to email broken@smallvictori.es if you need help :-)

domain: example.com
title: Title One!

# Any errors will appear here

test error

`

func TestSettings_TestGood(t *testing.T) {
	content, errs := parseSettings([]byte(exampleFile))
	if len(errs) != 0 {
		t.Fatalf("should not have err: %v", errs)
	}

	if content.Domain != "foo.example.com" {
		t.Fatalf("domain not expected: %s", content.Domain)
	}

	if content.Title != "Title One!" {
		t.Fatalf("title not expected: %s", content.Title)
	}

}

func TestSettings_TestBad(t *testing.T) {
	content, errs := parseSettings([]byte(badExampleFile))

	if len(errs) != 1 {
		t.Fatalf("should have err: %v", errs)
	}

	if content.Domain != "example.com" {
		t.Fatalf("domain not expected: %s", content.Domain)
	}

	if content.Title != "Title One!" {
		t.Fatalf("title not expected: %s", content.Title)
	}

}

func TestSettings_CreateSettings(t *testing.T) {
	var u = UserSettings{Title: "Title One!", Domain: "foo.example.com"}
	var noerrs = make([]error, 0)

	content, err := createSettings(u, noerrs)

	if err != nil {
		t.Fatalf("should not have err: %v", err)
	}

	if string(content) != string([]byte(exampleFile)) {
		t.Fatalf("created settings does not match: \n\n%s\n\n%s", content, []byte(exampleFile))
	}

}

func TestSettings_CreateSettingsWithErrors(t *testing.T) {
	var u = UserSettings{Title: "Title One!", Domain: "example.com"}
	var errs = make([]error, 0)
	errs = append(errs, errors.New("test error"))

	content, err := createSettings(u, errs)

	if err != nil {
		t.Fatalf("should not have err: %v", err)
	}

	if string(content) != string([]byte(badExampleFile)) {
		t.Fatalf("created settings with errs does not match: \n\n%s\n\n%s", content, []byte(badExampleFile))
	}

}
