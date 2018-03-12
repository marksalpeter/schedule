## schedule: A Golang Job Scheduling Package.
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/marksalpeter/schedule)
+[![Go Report Card](https://goreportcard.com/badge/github.com/marksalpeter/schedule)](https://goreportcard.com/report/github.com/marksalpeter/schedule)

schedule is a Golang job scheduling package which lets you run Go functions periodically at pre-determined interval using a simple, human-friendly syntax.

schedule can optionally use a mysql database to guarentee that each scheduled job only runs once per cluster of servers

schedule is inspired by the Ruby module [clockwork](<https://github.com/tomykaira/clockwork>) and Python job scheduling package [schedule](<https://github.com/dbader/schedule>)

This package has been heavily inspired by the good, but rather buggy [goCron](https://github.com/jasonlvhit/gocron) package

See also this two great articles:
* [Rethinking Cron](http://adam.herokuapp.com/past/2010/4/13/rethinking_cron/)
* [Replace Cron with Clockwork](http://adam.herokuapp.com/past/2010/6/30/replace_cron_with_clockwork/)

Back to this package, you could just use this simple API as below, to run a cron scheduler.

``` go
package main

import (
	"fmt"
	"github.com/marksalpeter/schedule"
)

func task(j Job, _ time.Time) {
	fmt.Printf("I run every %s %s\n", j.Amount(), j.Interval())
}

func main() {

	// add an example for every task
	now := time.Now()
	schedule.Add("once-task").Once().Starting(now).Do(task)
	schedule.Add("second-task").Every(1).Seconds().Starting(now).Do(task)
	schedule.Add("minute-task").Every(1).Minutes().Starting(now).Do(task)
	schedule.Add("hour-task").Every(1).Hours().Starting(now).Do(task)
	schedule.Add("day-task").Every(1).Days().At(now.Hours(), now.Minutes(), now.Seconds()).Starting(now).Do(task)
	schedule.Add("week-task").Every(1).Weeks().On(int(now.Weekday())).At(now.Hours(), now.Minutes(), now.Seconds()).Starting(now).Do(task)
	schedule.Add("month-task").Every(1).Months().In(now.Month()).On(now.Day()).At(now.Hours(), now.Minutes(), now.Seconds()).Starting(now).Do(task)
	schedule.Add("year-task").Every(1).Years().In(now.Month()).On(now.Day()).At(now.Hours(), now.Minutes(), now.Seconds()).Starting(now).Do(task)

	// you can see all of the jobs in the scheduler here
	fmt.Printf("%+v\n", schedule.List())

	select{}
}
```

## Roadmap

schedule is in beta, but the api is very unlikely to change. here is what is needed fully releasable version 1

- [ ] finish writing tests for all functionality of the non db synchronized local version
- [x] test for database synchronicity
- [x] bug fix for database synchronicity
- [ ] examples for godoc
- [x] basic example in README.md
- [ ] database example in README.md
- [ ] set up vendor


