## schedule: A Golang Job Scheduling Package.
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/marksalpeter/schedule)

schedule is a Golang job scheduling package which lets you run Go functions periodically at pre-determined interval using a simple, human-friendly syntax.

schedule can optionally use a mysql database to synchronize its jobscheduling across multiple server instances

schedule is inspired by the Ruby module [clockwork](<https://github.com/tomykaira/clockwork>) and Python job scheduling package [schedule](<https://github.com/dbader/schedule>)

This package has been heavily inspired by the good, but rather buggy [goCron](https://github.com/jasonlvhit/gocron) package

## Roadmap

schedule is in beta, but the api is very unlikely to change. here is what is needed fully releasable version 1

- [ ] finish tests for the non db synchronized local version
- [ ] test for database synchronicity
- [ ] examples for godoc
- [ ] examples in README.md
- [ ] set up vendor


