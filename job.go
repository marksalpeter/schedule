package schedule

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Job represents a task that is queued on the system at a certian time
type Job interface {
	// Name is the name of the job. It is unique to the scheduler that it is added to
	Name() string

	// Amount is the amount of some interval of time that will elapse between executions.
	// If there is only 1 execution of this task, it will be set to zero
	Amount() int

	// Interval is the interval of time that will elapse between executions
	Interval() IntervalType

	// Description is a plain english sentence that describes when this job is executed
	Description() string

	// Scheduler is the `Scheduler` that this job belongs to
	Scheduler() Scheduler

	// execute executes the job if it needs an execution
	execute(time.Time) bool
}

// Amount determines the amount of some interval of time that will elapse between executions
type Amount interface {
	Every(i ...int) Interval
	Once() Starting
}

// Interval determines the interval of time that will elapse between executions
type Interval interface {
	Years() Month
	Months() Day
	Weeks() Day
	Days() Time
	Hours() Starting
	Minutes() Starting
	Seconds() Starting
}

// Month adds the month to the job
type Month interface {
	In(time.Month) Day
}

// Day adds the day to the job
type Day interface {
	On(day int) Time
}

// Time sets the time that the job will execute
type Time interface {
	At(hours, minutes, seconds int) Starting
}

// Starting set the time we start counting
type Starting interface {
	Starting(time.Time) Task
}

// Task adds the func that will be executed by the `Scheduler`. It is the final step in the `Job` builder methods.
type Task interface {
	Do(func(Job, time.Time)) error
}

// IntervalType is a string representation of the interval chosen by the `Interval` interface
type IntervalType string

const (
	// Once happens one time only
	Once = IntervalType("once")

	// Years is set if `Interval.Years` is called
	Years = IntervalType("years")

	// Months is set if `Interval.Months` is called
	Months = IntervalType("months")

	// Weeks is set if `Interval.Weeks` is called
	Weeks = IntervalType("weeks")

	// Days is set if `Interval.Days` is called
	Days = IntervalType("days")

	// Hours is set if `Interval.Hours` is called
	Hours = IntervalType("hours")

	// Minutes is set if `Interval.MinutesDaily` is called
	Minutes = IntervalType("minutes")

	// Seconds is set if `Interval.Seconds` is called
	Seconds = IntervalType("seconds")
)

// Scan implements `sql.Scanner`
func (it *IntervalType) Scan(value interface{}) error {
	*it = IntervalType(value.([]byte))
	return nil
}

// Value implements the `driver.Valuer` interface
func (it IntervalType) Value() (driver.Value, error) {
	return string(it), nil
}

// job implements `Job`, `Interval`, `Increment`, `Month`, `Day`, `Time`, `Starting`, and `Task` interfaces
type job struct {
	JobName        string `sql:"index"`
	IntervalAmount int
	IntervalType   IntervalType
	Month          int
	Day            int
	Hour           int
	Minute         int
	Second         int
	JobDuration    time.Duration
	StartAt        time.Time
	LastRunAt      time.Time
	NextRunAt      time.Time
	do             func(Job, time.Time)
	scheduler      Scheduler
}

// TableName makes sure that we add this job to the right scheduler in the db
func (j *job) TableName() string {
	return j.scheduler.Name()
}

// Name is the name of the job. It is unique to the scheduler that it is added to
func (j *job) Name() string {
	return j.JobName
}

// Amount is the amount of some interval of time that will elapse between executions.
// If there is only 1 execution of this task, it will be set to zero
func (j *job) Amount() int {
	return j.IntervalAmount
}

// Interval is the interval of time that will elapse between executions
func (j *job) Interval() IntervalType {
	return j.IntervalType
}

// Description is a plain english sentence that describes when this job is executed
func (j *job) Description() string {
	// TODO: write something better than this
	return fmt.Sprintf("%+v", j)
}

func (j *job) Scheduler() Scheduler {
	return j.scheduler
}

func (j *job) Every(i ...int) Interval {
	if i == nil {
		j.IntervalAmount = 1
		return j
	} else if i[0] == 0 {
		panic("call `Interval.Once` instead")
	} else if i[0] < 0 {
		panic("Every expects a number greater than 0")
	}
	j.IntervalAmount = i[0]
	return j
}

func (j *job) Once() Starting {
	j.IntervalAmount = 0
	j.IntervalType = Once
	return j
}

func (j *job) Years() Month {
	j.IntervalType = Years
	return j
}

func (j *job) Months() Day {
	j.IntervalType = Months
	return j
}

func (j *job) Weeks() Day {
	j.IntervalType = Weeks
	return j
}

func (j *job) Days() Time {
	j.IntervalType = Days
	return j
}

func (j *job) Hours() Starting {
	j.IntervalType = Hours
	return j
}

func (j *job) Minutes() Starting {
	j.IntervalType = Minutes
	return j
}

func (j *job) Seconds() Starting {
	j.IntervalType = Seconds
	return j
}

func (j *job) In(month time.Month) Day {
	j.Month = int(month)
	return j
}

func (j *job) On(day int) Time {
	if j.IntervalType == Weeks && (day < 0 || day > 6) {
		panic("day must be a valid time.Weekday when scheduling a weekly task")
	}
	j.Day = day
	return j
}

func (j *job) At(hours int, minutes int, seconds int) Starting {
	j.Hour = hours
	j.Minute = minutes
	j.Second = seconds
	return j
}

func (j *job) Starting(t time.Time) Task {
	j.StartAt = t
	j.caclulateNextRunAt(t)
	return j
}

func (j *job) Do(do func(Job, time.Time)) error {
	j.do = do
	return j.scheduler.add(j)
}

// execute handles all job and scheduling based logic
func (j *job) execute(now time.Time) bool {
	if j.NextRunAt.After(now) {
		return false
	}
	j.LastRunAt = j.NextRunAt
	j.caclulateNextRunAt(now)
	if err := j.scheduler.update(j); err != nil {
		return false
	}
	j.do(j, now)
	return true
}

// caclulateNextRunAt determines `job.NextRunAt`
func (j *job) caclulateNextRunAt(now time.Time) {
	switch j.IntervalType {
	case Years:
		j.NextRunAt = time.Date(j.StartAt.Year(), time.Month(j.Month), j.Day, j.Hour, j.Minute, j.Second, j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.AddDate(j.IntervalAmount-1, 0, 0)
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.AddDate(j.IntervalAmount, 0, 0)
		}
	case Months:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.Day, j.Hour, j.Minute, j.Second, j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.AddDate(0, j.IntervalAmount-1, 0)
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.AddDate(0, j.IntervalAmount, 0)
		}
	case Weeks:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.StartAt.Day(), j.Hour, j.Minute, j.Second, j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.AddDate(0, 0, j.Day-int(j.StartAt.Weekday()))
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.AddDate(0, 0, j.IntervalAmount*7)
		}
	case Days:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.StartAt.Day(), j.Hour, j.Minute, j.Second, j.StartAt.Nanosecond(), j.StartAt.Location())
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.AddDate(0, 0, 1)
		}
	case Hours:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.StartAt.Day(), j.StartAt.Hour(), j.StartAt.Minute(), j.StartAt.Second(), j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.Add(time.Hour * time.Duration(j.IntervalAmount))
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.Add(time.Hour * time.Duration(j.IntervalAmount))
		}
	case Minutes:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.StartAt.Day(), j.StartAt.Hour(), j.StartAt.Minute(), j.StartAt.Second(), j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.Add(time.Minute * time.Duration(j.IntervalAmount))
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.Add(time.Minute * time.Duration(j.IntervalAmount))
		}
	case Seconds:
		j.NextRunAt = time.Date(j.StartAt.Year(), j.StartAt.Month(), j.StartAt.Day(), j.StartAt.Hour(), j.StartAt.Minute(), j.StartAt.Second(), j.StartAt.Nanosecond(), j.StartAt.Location())
		j.NextRunAt = j.NextRunAt.Add(time.Second * time.Duration(j.IntervalAmount))
		for j.NextRunAt.Before(now) {
			j.NextRunAt = j.NextRunAt.Add(time.Second * time.Duration(j.IntervalAmount))
		}
	default:
		panic(fmt.Errorf("increment type %s not implemented", j.IntervalType))
	}
}

// formatDay formats the day in `Job.Description`
func formatDay(d int) string {
	var format string
	switch d % 10 {
	case 0, 4, 5, 6, 7, 8, 9:
		format = "%dth"
	case 1:
		format = "%dst"
	case 2:
		format = "%dnd"
	case 3:
		format = "%drd"
	default:
		format = "%d"
	}
	return fmt.Sprintf(format, d)
}
