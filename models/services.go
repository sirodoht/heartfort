package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type ServicesConfig func(*Services) error

// WithGorm will open a GORM connection with the provided
// info and attach it to the Services type if there aren't
// any errors.
func WithGorm(connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open("postgres", connectionInfo)
		if err != nil {
			return err
		}
		s.DB = db
		return nil
	}
}

// WithLogMode will set the LogMode of the current GORM DB
// object associated with the Services type. It is assumed
// that the DB object will already exist and be initialized
// properly, so you will want to call something like
// WithGorm before this config function.
func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.DB.LogMode(mode)
		return nil
	}
}

// WithUser will use the existing GORM DB connection of the
// Services object along with the provided pepper and hmacKey
// to build and set a UserService.
func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.DB, pepper, hmacKey)
		return nil
	}
}

// WithJob will use the existing GORM DB connection of
// the Services object to build and set a JobService.
func WithJob() ServicesConfig {
	return func(s *Services) error {
		s.Job = NewJobService(s.DB)
		return nil
	}
}

// WithAssignment will use the existing GORM DB connection of
// the Services object to build and set a AssignmentService.
func WithAssignment() ServicesConfig {
	return func(s *Services) error {
		s.Assignment = NewAssignmentService(s.DB)
		return nil
	}
}

func WithMate() ServicesConfig {
	return func(s *Services) error {
		s.Mate = NewMateService(s.DB)
		return nil
	}
}

// NewServices now will accept a list of config functions to
// run. Each function will accept a pointer to the current
// Services object as its only argument and will edit that
// object inline and return an error if there is one. Once
// we have run all configs we will return the Services object.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type Services struct {
	Mate       MateService
	Assignment AssignmentService
	Job        JobService
	User       UserService
	DB         *gorm.DB
}

// Closes the database connection
func (s *Services) Close() error {
	return s.DB.Close()
}

// AutoMigrate will attempt to automatically migrate all tables
func (s *Services) AutoMigrate() error {
	return s.DB.AutoMigrate(&User{}, &Job{}, &Assignment{}, &pwReset{}, &Mate{}).Error
}

// DestructiveReset drops all tables and rebuilds them
func (s *Services) DestructiveReset() error {
	err := s.DB.DropTableIfExists(&User{}, &Job{}, &Assignment{}, &pwReset{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}
