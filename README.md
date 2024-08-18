# FishBB (Alpha)

Simple, sustainable communities. Minimalist bulliten board software.

Flagship instance at [fishbb.fishbb.org](https://fishbb.fishbb.org).

## Tech

go + sqlite

no javascript

12 dependencies, 2000 lines of code

## Running

```sh
go run main.go
```

This will setup a database with the admin user with username 'admin' and password 'admin'. You can also set a custom db path:

```sh
go build 
./fishbb -path foo.db
```

## Self-hosting

FishBB is designed to require a minimal amount of infrastructure and
maintenance burden for self-hosting. Please reach out to me [alex@alexwennerberg.com](mailto:alex@alexwennerberg.com) if you are interested in running your own instance!

All FishBB data is stored in a single sqlite file. HTML templates are embedded in the Go bindary.

FishBB is VERY early in development -- expect bugs and be very wary of sensitive data. Make sure to change the admin password away from default credentials.

### Configuration

Admin configuration is available at `/control`

See the comments in `config.go` for now (better documentation forthcoming)

### Google Signup

Your forum can optionally allow Google Signup. You will need to create an
[OAuth App](https://developers.google.com/identity/protocols/oauth2) on Google
and set the client ID and client secret in the configuration file.

## Contributing

The mailing list for FishBB is at https://lists.sr.ht/~aw/fishbb-devel

Feel free to use the [flagship instance](https://fishbb.fishbb.org) as well for project discussion and feedback!
