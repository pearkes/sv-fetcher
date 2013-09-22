package main

import (
	"bytes"
	"errors"
	"github.com/knieriem/markdown"
	"github.com/pearkes/Dropbox-Go/dropbox"
	"github.com/pearkes/sv-frontend/data"
	"log"
	"os"
	"strconv"
	"strings"
)

type fetcher struct {
	Session  dropbox.Session // The dropbox session
	Id       string          // the id of the user
	Contents []dropbox.Contents
	Settings UserSettings
	User     data.User
	Hash     string
}

func NewFetcher(user data.User) *fetcher {
	s := dropbox.Session{
		AppKey:     os.Getenv("DROPBOX_KEY"),
		AppSecret:  os.Getenv("DROPBOX_SECRET"),
		AccessType: "app_folder",
		Token:      user.DropboxToken,
	}

	fet := &fetcher{
		s,
		string(user.Id),
		nil,
		UserSettings{},
		user,
		"",
	}

	return fet
}

// lists the sandbox folder and sets the contents onto the fetcher
func (f *fetcher) listFolder() {
	u := dropbox.Uri{
		Root: "sandbox",
	}
	meta, err := dropbox.GetMetadata(f.Session, u, nil)

	if err != nil {
		log.Printf("receieved error trying to list folder: %s", err)
	}

	f.Contents = meta.Contents
	f.Hash = meta.Hash
}

// gets the settings from a dropbox file, parses it, returns it.
func (f *fetcher) checkSettings() error {
	u := dropbox.Uri{
		Root: "sandbox/_settings.txt",
	}

	file, meta, err := dropbox.GetFile(f.Session, u, nil)

	if err != nil {
		log.Printf("receieved error trying to retrieve settings file: %s", err)
	}

	set, errs := parseSettings(file)
	if set.Domain == "" {
		f.Settings.Domain = f.User.Name
	} else {
		f.Settings.Domain = set.Domain
	}
	if set.Title == "" {
		f.Settings.Title = "A Small Victory"
	} else {
		f.Settings.Title = set.Title
	}

	if meta.Revision == f.User.SettingsRev {
		return errors.New("Settings revision matches")
	}

	content, err := createSettings(f.Settings, errs)

	if err != nil {
		log.Printf("receieved error trying to generate settings file: %s", err)
		return err
	}

	meta, err = dropbox.UploadFile(f.Session, content, u, nil)

	f.Settings.Revision = meta.Revision

	if err != nil {
		log.Printf("receieved error trying to create settings file: %s", err)
		return err
	}

	return err
}

// evaluates the contents of a fetcher and returns an array of
// assets to render
func (f *fetcher) evalFiles() ([]Asset, string) {
	var assets = make([]Asset, 0)
	var directUrl = ""
	var content = ""

	for _, c := range f.Contents {
		// Use the path to parse the meta data
		filename, tag, order := parseMetaData(c.Path)

		contentNeeded, ext := checkContentNeeded(c.Path)
		// If we need the content, get it here.
		if contentNeeded == true {
			var err error
			content, err = f.retrieveContent(c.Path, ext)

			if err != nil {
				log.Printf("failed getting content for file: %s", err.Error())
				// Move on to the next one
				continue
			}
			// Skip settings
			if c.Path == "/_settings.txt" {
				continue
			}

			// Special index.html type
			if c.Path == "/index.html" {
				return assets, content
			}

			// Set this to be injected later
			c.MimeType = "text/inject"
		} else {
			directUrl = f.generateLink(c.Path)
		}

		newAsset := Asset{
			Url:      directUrl,
			Content:  content,
			Mime:     c.MimeType,
			Tag:      tag,
			Filename: filename,
			Order:    order,
		}

		assets = append(assets, newAsset)
	}
	return assets, ""
}

func (f *fetcher) retrieveContent(path string, extension string) (string, error) {
	u := dropbox.Uri{
		Root: "sandbox" + path,
	}

	file, _, err := dropbox.GetFile(f.Session, u, nil)
	if err != nil {
		return "", err
	}

	// Markdown check
	if extension == "md" {
		p := markdown.NewParser(&markdown.Extensions{Smart: true})
		buf := new(bytes.Buffer)
		p.Markdown(bytes.NewReader(file), markdown.ToHTML(buf))
		return buf.String(), nil
	} else {
		return string(file), nil
	}

}

// generates a link for a file and returns it
func (f *fetcher) generateLink(path string) string {
	u := dropbox.Uri{
		Root: "sandbox" + path,
	}
	// Get a "share" url so the ID is permanent
	url, err := dropbox.Share(f.Session, u, &dropbox.Parameters{ShortUrl: "false"})
	if err != nil {
		log.Printf("receieved error trying to generate link: %s", err)
	}
	// Whoops, permanent direct links!
	return strings.Replace(url.Url, "www.dropbox", "dl.dropboxusercontent", 1)
}

// Checks to see if content is required and returns a bool
func checkContentNeeded(path string) (bool, string) {
	var extension string
	md := strings.HasSuffix(path, ".md")
	markd := strings.HasSuffix(path, ".markdown")
	txt := strings.HasSuffix(path, ".txt")
	html := strings.HasSuffix(path, ".html")

	if (md || markd) == true {
		extension = "md"
	}

	if txt == true {
		extension = "txt"
	}

	if html == true {
		extension = "html"
	}

	return markd || md || txt || html || false, extension
}

// Takes a string path as a string and returns the tag and path and order
func parseMetaData(path string) (filename string, tag string, order int64) {
	filename = strings.TrimPrefix(path, "/")
	name := strings.Split(filename, ".")[0]
	parts := strings.Split(name, "_")
	tag = parts[0]
	if len(parts) == 2 {
		var err error
		order, err = strconv.ParseInt(parts[1], 8, 64)
		if err != nil {
			order = 1
		}
	} else {
		// no ordering
		order = 1
	}
	return
}
