package schedule_test

import (
	"testing"
	"time"

	"github.com/marksalpeter/schedule"
	"github.com/stretchr/testify/assert"
)

func TestSeconds(t *testing.T) {
	s := schedule.New(&schedule.Config{
		Name: "test",
	})
	now := time.Now()
	var amounts []int
	test := func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	}
	s.Add("1-second").Every(1).Seconds().Starting(now).Do(test)
	s.Add("2-second").Every(2).Seconds().Starting(now).Do(test)
	s.Add("3-second").Every(3).Seconds().Starting(now).Do(test)
	s.Add("4-second").Every(4).Seconds().Starting(now).Do(test)
	s.Add("5-second").Every(5).Seconds().Starting(now).Do(test)

	s.Start()
	<-time.NewTimer(10 * time.Second).C
	s.Stop()
	assert.New(t).Equal([]int{
		1,
		1, 2,
		1, 3,
		1, 2, 4,
		1, 5,
		1, 2, 3,
		1,
		1, 2, 4,
		1, 3,
		1, 2, 5,
	}, amounts, "the seconds are in the correct order")

}
func TestDatabaseSeconds(t *testing.T) {

	// create our test function and output collection
	var amounts []int
	test := func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	}

	// create 10 competing test schedulers
	var ss []schedule.Scheduler
	config := schedule.Config{
		Name:     "second-test-scheduler",
		Database: "test",
		Instance: "127.0.0.1:3306",
		Username: "test",
		Password: "test",
		// LogDB:    true,
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		s := schedule.New(&config)
		s.Add("1-second").Every(1).Seconds().Starting(now).Do(test)
		s.Add("2-second").Every(2).Seconds().Starting(now).Do(test)
		s.Add("3-second").Every(3).Seconds().Starting(now).Do(test)
		s.Add("4-second").Every(4).Seconds().Starting(now).Do(test)
		s.Add("5-second").Every(5).Seconds().Starting(now).Do(test)
		s.Start()
		ss = append(ss, s)
	}

	// wait 10 seconds to collect the output
	<-time.NewTimer(10 * time.Second).C
	for _, s := range ss {
		s.Stop()
	}
	assert.New(t).Equal([]int{
		1,
		1, 2,
		1, 3,
		1, 2, 4,
		1, 5,
		1, 2, 3,
		1,
		1, 2, 4,
		1, 3,
		1, 2, 5,
	}, amounts, "the seconds are in the correct order")
}

func TestDatabaseOnce(t *testing.T) {

	// create our test function and output collection
	var amounts []int
	test := func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	}

	// create 10 competing test schedulers
	var ss []schedule.Scheduler
	config := schedule.Config{
		Name:     "second-test-scheduler",
		Database: "test",
		Instance: "127.0.0.1:3306",
		Username: "test",
		Password: "test",
		// LogDB:    true,
	}
	now := time.Now().Add(5 * time.Second)
	for i := 0; i < 10; i++ {
		s := schedule.New(&config)
		s.Add("once").Once().Starting(now).Do(test)
		s.Start()
		ss = append(ss, s)
	}

	// wait 10 seconds to collect the output
	<-time.NewTimer(10 * time.Second).C
	for _, s := range ss {
		s.Stop()
	}
	assert.New(t).Equal([]int{
		0,
	}, amounts, "the seconds are in the correct order")
}
