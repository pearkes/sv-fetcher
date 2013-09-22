package main

import (
	"testing"
)

func TestRender_TestEvalAssets(t *testing.T) {
	asset1 := Asset{
		Url:      "http://dropbox.com/thing",
		Content:  "content",
		Mime:     "text/inject",
		Tag:      "winter",
		Filename: "winter.txt",
		Order:    1,
	}

	asset2 := Asset{
		Url:      "http://dropbox.com/thing.css",
		Content:  "",
		Mime:     "text/css",
		Tag:      "winter",
		Filename: "winter_02.css",
		Order:    2,
	}

	asset3 := Asset{
		Url:      "http://dropbox.com/thing.png",
		Content:  "",
		Mime:     "application/javascript",
		Tag:      "summer",
		Filename: "summer.js",
		Order:    2,
	}

	asset4 := Asset{
		Url:      "http://dropbox.com/thing.png",
		Content:  "",
		Mime:     "image/png",
		Tag:      "fall",
		Filename: "fall.png",
		Order:    1,
	}

	asset5 := Asset{
		Url:      "http://dropbox.com/thing.png",
		Content:  "",
		Mime:     "image/png",
		Tag:      "fall",
		Filename: "fall_2.png",
		Order:    2,
	}

	asset6 := Asset{
		Url:      "http://dropbox.com/thing.png",
		Content:  "",
		Mime:     "image/png",
		Tag:      "fall",
		Filename: "fall_3.png",
		Order:    3,
	}

	assets := []Asset{asset1, asset2, asset6, asset3, asset4, asset5}

	page := evalAssets(assets)

	if page.Title != "" {
		t.Fatal("title should be blank")
	}

	if len(page.Projects) != 2 {
		t.Fatalf("page should have 2 projects: %v", page)
	}

	if len(page.Projects["winter"].Assets) != 1 {
		t.Fatal("incorrect number of winter assets")
	}

	if page.Projects["winter"].Assets[0].Content != "content" {
		t.Fatal("content is incorrect in text file")
	}

	if len(page.Projects["summer"].Assets) != 0 {
		t.Fatal("incorrect number of summer assets")
	}

	if len(page.Js) != 1 {
		t.Fatal("incorrect number of javascript files")
	}

	if len(page.Css) != 1 {
		t.Fatal("incorrect number of css files")
	}

	if len(page.Projects["fall"].Assets) != 3 {
		t.Fatal("incorrect number of fall assets")
	}

	if page.Projects["fall"].Assets[0].Order != 1 {
		t.Fatal("incorrect ordering")
	}

	if page.Projects["fall"].Assets[1].Order != 2 {
		t.Fatal("incorrect ordering")
	}

	if page.Projects["fall"].Assets[2].Order != 3 {
		t.Fatal("incorrect ordering")
	}
}
