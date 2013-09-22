package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var temps = template.Must(template.ParseFiles("files/settings.txt"))

type SettingsTemplate struct {
	Errs   []error
	Title  string
	Domain string
}

type UserSettings struct {
	Title    string // the title of their site
	Domain   string // the domain name for their site
	Revision int    // the dropbox revision for the settings
}

// Parses a bite-array settings file into a UserSettings object
func parseSettings(content []byte) (UserSettings, []error) {
	var errs = make([]error, 0)
	var settings UserSettings
	lines := strings.Split(string(content), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "domain:") {
			u := strings.Split(l, ":")[1]
			u = strings.TrimSpace(u) // trim leading and trailing whitespace
			punkts := strings.Split(u, ".")
			if len(punkts) != 3 {
				log.Printf("failed to parse url: %s", u)
				errs = append(errs, errors.New("we couldn't parse and update your domain, try again"))
			}
			settings.Domain = u
		}
		if strings.HasPrefix(l, "title:") {
			t := strings.Split(l, ":")[1]
			t = strings.TrimSpace(t) // trim leading and trailing whitespace
			settings.Title = t
		}
	}
	return settings, errs
}

// Creates a settings file and returns a []byte for writing to dropbox
func createSettings(settings UserSettings, errs []error) ([]byte, error) {
	p := &SettingsTemplate{Errs: errs, Title: settings.Title, Domain: settings.Domain}
	buf := new(bytes.Buffer)
	err := temps.ExecuteTemplate(buf, "settings.txt", p)
	return buf.Bytes(), err
}

func herokuDomainCreate(domain string) error {
	client := &http.Client{}
	body := make(map[string]interface{})
	body["hostname"] = domain
	jbod, _ := json.Marshal(body)
	buf := new(bytes.Buffer)
	buf.Write(jbod)
	herokuUrl := fmt.Sprintf("https://api.heroku.com/apps/%s/domains", os.Getenv("HEROKU_APP"))
	req, _ := http.NewRequest("POST", herokuUrl, buf)
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("HEROKU_USER"), os.Getenv("HEROKU_TOKEN"))
	resp, err := client.Do(req)
	log.Printf("Response for Heroku domain create: %v", resp.StatusCode)
	return err
}
