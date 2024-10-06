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

Tagged versions should be stable. Main branch is not guaranteed to be.

FishBB is designed to require a minimal amount of infrastructure and
maintenance burden for self-hosting. Please reach out to me [alex@alexwennerberg.com](mailto:alex@alexwennerberg.com) if you are interested in running your own instance!

All FishBB data is stored in a single sqlite file. HTML templates are embedded in the Go bindary.

FishBB is VERY early in development -- expect bugs and be very wary of sensitive data. Make sure to change the admin password away from default credentials.

## As a package

FishBB also runs as a 'cluster' where multiple forums can be created. These forums can either exist in the 'cluster' (as a sort of broader forum manager) or exported on their own.

The app that hosts the fishbb cluster is https://git.sr.ht/~aw/fishbb-cluster

### Configuration

Admin configuration is available at `/control`

See the comments in `config.go` for now (better documentation forthcoming)

## Contributing

The mailing list for FishBB is at https://lists.sr.ht/~aw/fishbb-devel

Feel free to use the [flagship instance](https://fishbb.fishbb.org) as well for project discussion and feedback!

FishBB is free software, if you'd like to, please <a href="https://www.patreon.com/alexwennerberg">donate</a> to support development
