# FishBB

Simple forum software (WIP)

flagship instance at https://fishbb.fishbb.org

go + sqlite

no javascript

dual hosted on https://git.sr.ht/~aw/fisbb and https://github.com/fishbb

## Running

`go run main.go`

This will setup a database with the admin user of "admin/admin"

## Self-hosting

FishBB is intended to require a minimal amount of infrastructure and
administration for self-hosting. Please reach out if you are interested in
running your own instance!

All fishBB data is stored in a single sqlite file, configured by -path

Running FishBB for the first time will create a database file and an admin user 
with the credentials "admin/admin".

### Configuration

Admin configuration is available at `/control`

See the comments in `config.go` for now (better documentation forthcoming)

TODO

**robots.txt** -- edit the views/robots.txt file, or, if you're running this service behind a reverse proxy, use that as well.


### Google Signup

Optional Google Signup TODO

If you are following along, anti-spam is NOT to the point where open signups
are a good idea. Don't try to self-host with open signups yet
