package models

import "github.com/jinzhu/gorm"

type Mate struct {
	gorm.Model
	Email string `gorm:"not_null"`
}

func NewMateService(db *gorm.DB) MateService {
	return &mateService{
		MateDB: &mateValidator{
			MateDB: &mateGorm{
				db: db,
			},
		},
	}
}

type MateService interface {
	MateDB
}

type mateService struct {
	MateDB
}

// MateDB is used to interact with the mates database.
type MateDB interface {
	ByID(id uint) (*Mate, error)
	List() ([]Mate, error)
	Create(mate *Mate) error
	Update(mate *Mate) error
	Delete(id uint) error
}

type mateValidator struct {
	MateDB
}

func (mv *mateValidator) Create(mate *Mate) error {
	err := runMateValFns(mate, mv.emailRequired)
	if err != nil {
		return err
	}
	return mv.MateDB.Create(mate)
}

func (mv *mateValidator) Update(mate *Mate) error {
	err := runMateValFns(mate, mv.emailRequired)
	if err != nil {
		return err
	}
	return mv.MateDB.Update(mate)
}

func (mv *mateValidator) Delete(id uint) error {
	var mate Mate
	mate.ID = id
	if err := runMateValFns(&mate, mv.nonZeroID); err != nil {
		return err
	}
	return mv.MateDB.Delete(mate.ID)
}

var _ MateDB = &mateGorm{}

type mateGorm struct {
	db *gorm.DB
}

func (jg *mateGorm) ByID(id uint) (*Mate, error) {
	var mate Mate
	db := jg.db.Where("id = ?", id)
	err := first(db, &mate)
	if err != nil {
		return nil, err
	}
	return &mate, nil
}

func (jg *mateGorm) List() ([]Mate, error) {
	var mates []Mate
	if err := jg.db.Find(&mates).Error; err != nil {
		return nil, err
	}
	return mates, nil
}

func (jg *mateGorm) Create(mate *Mate) error {
	return jg.db.Create(mate).Error
}

func (jg *mateGorm) Update(mate *Mate) error {
	return jg.db.Save(mate).Error
}

func (jg *mateGorm) Delete(id uint) error {
	mate := Mate{Model: gorm.Model{ID: id}}
	return jg.db.Delete(&mate).Error
}

func (mv *mateValidator) emailRequired(mate *Mate) error {
	if mate.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (mv *mateValidator) nonZeroID(mate *Mate) error {
	if mate.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

type mateValFn func(*Mate) error

func runMateValFns(mate *Mate, fns ...mateValFn) error {
	for _, fn := range fns {
		if err := fn(mate); err != nil {
			return err
		}
	}
	return nil
}
