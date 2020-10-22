package models

import "github.com/jinzhu/gorm"

const (
	ErrNameRequired modelError = "models: name is required"
)

// Job represents the jobs table in our DB and is a single job
// such as "Kitchen".
type Job struct {
	gorm.Model
	Name string `gorm:"not_null"`
}

func NewJobService(db *gorm.DB) JobService {
	return &jobService{
		JobDB: &jobValidator{
			JobDB: &jobGorm{
				db: db,
			},
		},
	}
}

type JobService interface {
	JobDB
}

type jobService struct {
	JobDB
}

// JobDB is used to interact with the jobs database.
type JobDB interface {
	ByID(id uint) (*Job, error)
	List() ([]Job, error)
	Create(job *Job) error
	Update(job *Job) error
	Delete(id uint) error
}

type jobValidator struct {
	JobDB
}

func (jv *jobValidator) Create(job *Job) error {
	err := runJobValFns(job, jv.nameRequired)
	if err != nil {
		return err
	}
	return jv.JobDB.Create(job)
}

func (jv *jobValidator) Update(job *Job) error {
	err := runJobValFns(job, jv.nameRequired)
	if err != nil {
		return err
	}
	return jv.JobDB.Update(job)
}

func (jv *jobValidator) Delete(id uint) error {
	var job Job
	job.ID = id
	if err := runJobValFns(&job, jv.nonZeroID); err != nil {
		return err
	}
	return jv.JobDB.Delete(job.ID)
}

var _ JobDB = &jobGorm{}

type jobGorm struct {
	db *gorm.DB
}

func (jg *jobGorm) ByID(id uint) (*Job, error) {
	var job Job
	db := jg.db.Where("id = ?", id)
	err := first(db, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (jg *jobGorm) List() ([]Job, error) {
	var jobs []Job
	if err := jg.db.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (jg *jobGorm) Create(job *Job) error {
	return jg.db.Create(job).Error
}

func (jg *jobGorm) Update(job *Job) error {
	return jg.db.Save(job).Error
}

func (jg *jobGorm) Delete(id uint) error {
	job := Job{Model: gorm.Model{ID: id}}
	return jg.db.Delete(&job).Error
}

func (jv *jobValidator) nameRequired(g *Job) error {
	if g.Name == "" {
		return ErrNameRequired
	}
	return nil
}

func (jv *jobValidator) nonZeroID(job *Job) error {
	if job.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

type jobValFn func(*Job) error

func runJobValFns(job *Job, fns ...jobValFn) error {
	for _, fn := range fns {
		if err := fn(job); err != nil {
			return err
		}
	}
	return nil
}
