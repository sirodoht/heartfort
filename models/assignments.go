package models

import "github.com/jinzhu/gorm"

const (
	ErrJobIDRequired        modelError = "models: job ID is required"
	ErrAssignmentIDRequired modelError = "models: assignment ID is required"
)

// Assignment represents the assignments table in our DB and is
// when a user is assigned to a assignment.
type Assignment struct {
	gorm.Model
	UserID uint `gorm:"not_null;index"`
	JobID  uint `gorm:"not_null"`
}

func NewAssignmentService(db *gorm.DB) AssignmentService {
	return &assignmentService{
		AssignmentDB: &assignmentValidator{
			AssignmentDB: &assignmentGorm{
				db: db,
			},
		},
	}
}

type AssignmentService interface {
	AssignmentDB
}

type assignmentService struct {
	AssignmentDB
}

// AssignmentDB is used to interact with the assignments database.
type AssignmentDB interface {
	ByID(id uint) (*Assignment, error)
	ByUserID(userID uint) ([]Assignment, error)
	List() ([]Assignment, error)
	Create(assignment *Assignment) error
	Update(assignment *Assignment) error
	Delete(id uint) error
}

type assignmentValidator struct {
	AssignmentDB
}

func (av *assignmentValidator) Create(assignment *Assignment) error {
	err := runAssignmentValFns(assignment, av.userIDRequired, av.jobIDRequired)
	if err != nil {
		return err
	}
	return av.AssignmentDB.Create(assignment)
}

func (av *assignmentValidator) Update(assignment *Assignment) error {
	err := runAssignmentValFns(assignment, av.userIDRequired, av.jobIDRequired)
	if err != nil {
		return err
	}
	return av.AssignmentDB.Update(assignment)
}

func (av *assignmentValidator) Delete(id uint) error {
	var assignment Assignment
	assignment.ID = id
	if err := runAssignmentValFns(&assignment, av.nonZeroID); err != nil {
		return err
	}
	return av.AssignmentDB.Delete(assignment.ID)
}

var _ AssignmentDB = &assignmentGorm{}

type assignmentGorm struct {
	db *gorm.DB
}

func (ag *assignmentGorm) ByID(id uint) (*Assignment, error) {
	var assignment Assignment
	db := ag.db.Where("id = ?", id)
	err := first(db, &assignment)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (ag *assignmentGorm) ByUserID(userID uint) ([]Assignment, error) {
	var assignments []Assignment
	db := ag.db.Where("user_id = ?", userID)
	if err := db.Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (ag *assignmentGorm) List() ([]Assignment, error) {
	var assignments []Assignment
	if err := ag.db.Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (ag *assignmentGorm) Create(assignment *Assignment) error {
	return ag.db.Create(assignment).Error
}

func (ag *assignmentGorm) Update(assignment *Assignment) error {
	return ag.db.Save(assignment).Error
}

func (ag *assignmentGorm) Delete(id uint) error {
	assignment := Assignment{Model: gorm.Model{ID: id}}
	return ag.db.Delete(&assignment).Error
}

func (av *assignmentValidator) userIDRequired(a *Assignment) error {
	if a.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (av *assignmentValidator) jobIDRequired(a *Assignment) error {
	if a.JobID <= 0 {
		return ErrJobIDRequired
	}
	return nil
}

func (av *assignmentValidator) nonZeroID(assignment *Assignment) error {
	if assignment.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

type assignmentValFn func(*Assignment) error

func runAssignmentValFns(assignment *Assignment, fns ...assignmentValFn) error {
	for _, fn := range fns {
		if err := fn(assignment); err != nil {
			return err
		}
	}
	return nil
}
