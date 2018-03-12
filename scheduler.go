// Package schedule is a golang job scheduling package
//
// Schedule is a Golang job scheduling package which lets you run Go functions periodically at pre-determined interval using a simple, human-friendly syntax.
// Schedule can optionally use a mysql database to synchronize its jobscheduling across multiple server instances.
// Schedule is inspired by the Ruby module [clockwork](<https://github.com/tomykaira/clockwork>) and Python job scheduling package [schedule](<https://github.com/dbader/schedule>).
// This package has been heavily inspired by the good, but rather buggy [goCron](https://github.com/jasonlvhit/gocron) package.
//
package schedule

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // import the sql driver
)

// Scheduler executes a sets of `Jobs` at a given time
type Scheduler interface {
	// Name is the unique name of the scheduler. Note: any scheduler with the same name will reference the same table name for synchronicity purposes
	Name() string

	// List returs a list of jobs added to this scheduler
	List() []Job

	// Add create a new job ascociated with the scheduler and returns its first builder method
	// Note: it will not be added to the scheduler until it is done being built (ie `Do` is called)
	Add(name string) Amount

	// Start starts the scheduler
	Start()

	// Stop stops the scheduler
	Stop()

	// add is used by the job to add itsself to the scheduler after it is done being built (ie `Do` is called).
	// It will optionally also be added to the database depending on how the scheduler is configured
	add(j *job) error

	// update checks the `NextRunAt` field in a synchronous way in the database to determine if
	// if it returns an error, the job should not be executed
	update(j *job) error
}

// Config configures the scheduler
type Config struct {
	// Name is the name of the scheduler
	Name string

	// Database is the name of the mysql database used to synchronize the scheduler
	// If a database is not passed in, the scheduler will not use database synchronicity
	Database string

	// Instancs is the address of the database instance used to synchronize the scheduler
	Instance string

	// Username is the username of the mysql user
	Username string

	// Password is the password of the mysql user
	Password string

	// LogDB when set to true, all sql transactions will be logged
	LogDB bool
}

// New creates a new `Scheduler`
func New(cfg *Config) Scheduler {
	// create the scheduler
	var s scheduler
	s.name = cfg.Name

	// open the database
	if len(cfg.Database) > 0 {
		db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Instance, cfg.Database))
		if err != nil {
			panic(err)
		}
		db.SingularTable(true)
		db.LogMode(cfg.LogDB)
		if err := db.AutoMigrate(&job{
			scheduler: &s,
		}).Error; err != nil {
			panic(err)
		}
		s.db = db
	}

	return &s
}

// DefaultScheduler is the `Scheduler`` referenced by the `Add` and `List` funcs
var DefaultScheduler = New(&Config{Name: "default"})

func init() {
	DefaultScheduler.Start()
}

// Add adds jobs to the `DefaultScheduler`
func Add(name string) Amount {
	return DefaultScheduler.Add(name)
}

// List returns the jobs from the `DefaultScheuler`
func List() []Job {
	return DefaultScheduler.List()
}

// scheduler implments `Scheduler`
type scheduler struct {
	name string
	jobs []Job
	db   *gorm.DB
	quit chan struct{}
	done chan struct{}
}

// Name is the unique name of the scheduler. Note: any scheduler with the same name will reference the same table name for synchronicity purposes
func (s *scheduler) Name() string {
	return s.name
}

// List returs a list of jobs added to this scheduler
func (s *scheduler) List() []Job {
	return s.jobs
}

// Add create a new job ascociated with the scheduler and returns its first builder method
// Note: it will not be added to the scheduler until it is done being built (ie `Do` is called)
func (s *scheduler) Add(name string) Amount {
	var j job
	j.JobName = name
	j.scheduler = s
	return &j
}

// Start starts the scheduler
func (s *scheduler) Start() {
	// stop the ticker
	if s.quit != nil {
		s.Stop()
	}

	// start the ticker
	s.quit = make(chan struct{})
	s.done = make(chan struct{})
	started := make(chan struct{})
	go func(s *scheduler, started chan struct{}) {
		ticker := time.NewTicker(time.Second)
		close(started)
		for {
			select {
			case t := <-ticker.C:
				for _, j := range s.jobs {
					j.execute(t)
				}
				break
			case <-s.quit:
				ticker.Stop()
				close(s.done)
				return
			}
		}
	}(s, started)
	<-started
}

// Stop stops the scheduler
func (s *scheduler) Stop() {
	if s.quit == nil {
		return
	}
	close(s.quit)
	<-s.done
	s.quit = nil
	s.done = nil
}

// add is used by the job to add itsself to the scheduler after it is done being built (ie `Do` is called).
// It will optionally also be added to the database depending on how the scheduler is configured
func (s *scheduler) add(j *job) error {
	for _, a := range s.jobs {
		if a.Name() == j.Name() {
			return fmt.Errorf("%s is already added to the scheduler", j.Name())
		}
	}

	// don't forget to append the job to the list of jobs in the scheduler at the end of this
	defer func() {
		s.jobs = append(s.jobs, j)
	}()

	// no database logic needed
	if s.db == nil {
		return nil
	}

	// select the job from the database
	tx := s.db.Begin()
	var dbJ job
	if err := tx.Raw(fmt.Sprintf("select * from `%s` where `job_name` = \"%s\" for update", s.name, j.JobName)).Scan(&dbJ).Error; err == gorm.ErrRecordNotFound {
		// create a new job in the database
		log.Println("CREATE")
		if err := tx.Create(j).Error; err != nil {
			if err := tx.Rollback().Error; err != nil {
				log.Println(err)
				return nil
			}
			log.Println(err)
			return nil
		}

	} else if err != nil {
		// catasriphic server error
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	} else if err := tx.Save(j).Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	}
	// commit the change to the db
	if err := tx.Commit().Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		log.Println(err)
	}
	return nil
}

// update checks the `NextRunAt` field in a synchronous way in the database to determine if
// if it returns an error, the job should not be executed
func (s *scheduler) update(j *job) error {
	if s.db == nil {
		return nil
	}
	var dbJ job
	tx := s.db.Begin()
	if err := tx.Raw(fmt.Sprintf("select * from `%s` where `job_name` = \"%s\" for update", s.name, j.JobName)).Scan(&dbJ).Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	}
	// check to see if another instance using the same database aready performed this execution
	if dbJ.NextRunAt.After(j.NextRunAt) || dbJ.NextRunAt.Equal(j.NextRunAt) {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return fmt.Errorf("another instance already executed")
	}
	// save our new run info
	if err := tx.Save(j).Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	}
	// commit the change to the db
	if err := tx.Commit().Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
	}
	return nil
}
