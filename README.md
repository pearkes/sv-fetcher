<p align="center">
    <a href="https://smallvictori.es"><img src="https://f.cloud.github.com/assets/846194/1235472/24f70d94-29a7-11e3-835a-84f55972b657.png" /></a>
</p>

This is service for [Small Victories](https://smallvictori.es) that
continually polls a user's Dropbox for changes. If changes are detected,
it renders the assets in the folder into a HTML webpage and places
the page in a cache for the [frontend](https://github.com/pearkes/sv-frontend)
to retrieve and display.

## Design

In order to limit load on Dropbox, the folder revision is saved after
each update, and updates are then only made if the folder revision has
changed.

The service is designed to run continually on Heroku on a single "dyno",
sleeping between API calls to limit load on Dropbox.

Currently, it processes about 1000 users every 10 seconds, assuming
many of the users revisions have not changed. If the userbase
were to grow far beyond that further load limiting functionality would
likely be introduced.

Considerations for the design were as follows:

- Keep cost of running continually at free, or close to free, if possible
- Speed of updates after a user makes changes
- Limiting abuse of Dropbox's API

## Hacking and Deploying

This service is written entirely in [Go](http://golang.org/), so you'll need to download
and install that.

You'll also need [Foreman](http://ddollar.github.io/foreman/) to start the application with the proper environment
variables.

To configure the service, you can use the configuration example and create
a `.env` file containing your credentials.

To run, you then use foreman:

    $ go build
    $ foreman run ./sv-fetcher
    ...

To run the tests:

    $ go test
    ...

To deploy:

    $ heroku create -b https://github.com/kr/heroku-buildpack-go.git
    ...
    $ git push heroku master
    ...

You'll need:

- Heroku account
- Heroku Postgres database
- Redis add-on of some kind
- Librato Metrics account

Use `heroku config:add` to add environment variables based on your
various tokens from the above to create the production environment,
as it is described in `.env.example`.

### Shared Credentials

Keep in mind these credentials should be shared with the [sv-frontend](https://github.com/pearkes/sv-frontend),
so you'll need to add the same environment variables there.

## Contributions

Small Victories being open source is mostly educational, as there is
unknown intent for further development.

If you're interested in maintaining or contributing to the project and
website, please contact us and we'll chat about it. Thanks!

[computers@smallvictori.es](mailto:computers@smallvictori.es)

## License

See [license](LICENSE.md) file.
