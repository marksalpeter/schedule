package schedule_test

import (
	"testing"
	"time"

	"github.com/marksalpeter/schedule"
	"github.com/stretchr/testify/assert"
)

func TestSecondsScheduler(t *testing.T) {
	s := schedule.New(&schedule.Config{
		Name: "Test",
	})
	now := time.Now()
	var amounts []int
	s.Add("1-second-test").Every(1).Seconds().Starting(now).Do(func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	})
	s.Add("2-second-test").Every(2).Seconds().Starting(now).Do(func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	})
	s.Add("3-second-test").Every(3).Seconds().Starting(now).Do(func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	})
	s.Add("4-second-test").Every(4).Seconds().Starting(now).Do(func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	})
	s.Add("5-second-test").Every(5).Seconds().Starting(now).Do(func(j schedule.Job, now time.Time) {
		amounts = append(amounts, j.Amount())
	})

	done := s.Start()
	timeout := time.NewTimer(10 * time.Second).C
	for {
		select {
		case <-done:
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
			return
		case <-timeout:
			s.Stop()
		}
	}
}
