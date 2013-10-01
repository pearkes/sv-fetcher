package main

import (
	"github.com/pearkes/sv-frontend/data"
	"github.com/pearkes/sv-frontend/stats"
	"log"
	"os"
	"sync"
	"time"
)

const (
	RUN_DELAY      = 3  // how many seconds to sleep between runs
	PER_USER_DELAY = 50 // how many milleseconds to wait between users
)

var db *data.Orm = nil
var r *data.Red = nil
var metrics *stats.StatsSink = nil

// Configure the various services and start the update loop for the users
// dropbox connections.
func main() {
	db = data.NewOrm(os.Getenv("DATABASE_CONNECTION"))
	r = data.NewRedis(os.Getenv("REDIS_ADDRESS"), os.Getenv("REDIS_AUTH"))
	metrics = stats.NewStatsSink(os.Getenv("LIBRATO_USER"), os.Getenv("LIBRATO_TOKEN"), stats.ENV_WORKER)

	for {
		log.Printf("GOD: Starting run")
		// Retrieve the number of users for debugging
		count := db.UserCount()
		log.Printf("GOD: Number of users: %v", count)

		// Keep track about how many jobs we queued
		processed := 0

		// retrieve all users from the datbase
		var users []data.User
		err := db.Hd.Where("id", ">", 0).Find(&users)

		if err != nil {
			log.Printf("Error retrieving users: %s", err.Error())
		}

		var wg sync.WaitGroup

		for _, u := range users {
			// Increment the wait group
			wg.Add(1)
			// Increment the total processed users
			processed = processed + 1
			// Sleep between queing user sync to lower load on Dropbox API
			time.Sleep(PER_USER_DELAY * time.Millisecond)

			// Asynchronously check the users dropbox folder and save
			// the changes (if any) to their site.
			go func(u data.User) {
				// finish this user in the waitgroup on completion
				defer wg.Done()

				// Creates a fetcher, which talks to Dropbox
				fetcher := NewFetcher(u)
				log.Printf("WORKER (%v): Starting", u.Id)

				// Retrieve the _settings.txt file from the users dropbox
				// and update accordingly
				err = fetcher.checkSettings()

				// If settings were checked succesfully, update the users name
				if err == nil {
					err := db.UpdateName(u, fetcher.Settings.Domain, fetcher.Settings.Revision)

					// If the domain name update failed, log, otherwise,
					// update the domain with Heroku so routing functions.
					if err != nil {
						log.Printf("failed to update user name (domain): %s", err)
					} else {
						err = herokuDomainCreate(fetcher.Settings.Domain)
						if err != nil {
							log.Printf("failed to update domain name on heroku: %s", err)
						} else {
							log.Printf("WORKER (%v): User domain updated succesfully", u.Id)
						}
					}
				} else {
					// Catch errors for the settings check
					log.Printf("WORKER (%v): %s", u.Id, err.Error())
				}

				// Retrieves a list of the users folder and stores it in
				// the fetcher
				fetcher.listFolder()

				// If the folders revision has is the same (nothing has changed)
				// we can safely stop here, incrementing our stats and moving on.
				if fetcher.Hash == u.FolderSum {
					metrics.Event(stats.USER_PROCESSED)
					log.Printf("WORKER (%v): Folder sum matches, skipping checks", u.Id)
					return
				}

				log.Printf("WORKER (%v): Retrieved files: %v", u.Id, len(fetcher.Contents))
				// Fetch assets in the folder and evaluate them
				assets, indexSpecial := fetcher.evalFiles()

				// We have a easter egg for "index.html" which automatically
				// overrides are custom built index.html for power users. In this
				// case, we save their index.html directly to the redis cache
				if indexSpecial != "" {
					savePage(indexSpecial, u)
					metrics.Event(stats.USER_PROCESSED)
					log.Printf("WORKER (%v): User page rendered succesfully (with index)", u.Id)
					return
				}

				log.Printf("WORKER (%v): Evaluated assets: %v", u.Id, len(assets))
				// Evaluate the assets retrieved, i.e Markdown formatting
				page := evalAssets(assets)

				// Put the title on the page from the user settings
				page.Title = fetcher.Settings.Title
				renderedPage, err := renderPage(page)

				// If we've gotten this far, check for errs
				if err != nil {
					log.Printf("WORKER (%v): Error rendering template: %s", u.Id, err)
					metrics.Event(stats.PAGE_RENDER_ERROR)
					return
				} else {
					// Only save if the render worked
					err = savePage(renderedPage, u)
					if err != nil {
						log.Printf("WORKER (%v): Error saving page to redis: %s", u.Id, err)
						metrics.Event(stats.PAGE_RENDER_ERROR)
						return
					}
					metrics.Event(stats.USER_PROCESSED)
					log.Printf("WORKER (%v): User page rendered succesfully", u.Id)
				}

				// Update the revision has on our side
				err = db.UpdateSum(u, fetcher.Hash)
				if err != nil {
					log.Printf("WORKER (%v): Error update folder checksum: %s", u.Id, err)
				}
				log.Printf("WORKER (%v): Updated folder checksum: %s", u.Id, fetcher.Hash)
			}(u)
		}

		// Wait for all of the user fetches and updates to complete
		wg.Wait()

		log.Printf("GOD: Run complete, users proccessed: %v", processed)
		metrics.Event(stats.RUN_COMPLETE)

		// Send a total user count
		metrics.Raw(db.UserCount(), "_total_users_created")

		// Sleep arbitrarily between runs
		time.Sleep(time.Second * RUN_DELAY)
	}
}
