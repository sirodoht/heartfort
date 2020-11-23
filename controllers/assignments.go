package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/sirodoht/heartfort/models"
	"github.com/sirodoht/heartfort/views"
)

const (
	IndexAssignments = "index_assignments"
	ShowAssignment   = "show_assignment"
	EditAssignment   = "edit_assignment"
)

func NewAssignments(as models.AssignmentService, js models.JobService, db *gorm.DB, r *mux.Router) *Assignments {
	return &Assignments{
		New:       views.NewView("layout", "assignments/new"),
		ShowView:  views.NewView("layout", "assignments/show"),
		EditView:  views.NewView("layout", "assignments/edit"),
		IndexView: views.NewView("layout", "assignments/index"),
		as:        as,
		js:        js,
		db:        db,
		r:         r,
	}
}

type Assignments struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	as        models.AssignmentService
	js        models.JobService
	db        *gorm.DB
	r         *mux.Router
}

type AssignmentForm struct {
	UserID uint `schema:"user_id"`
	JobID  uint `schema:"job_id"`
}

// GET /assignments
func (a *Assignments) Index(w http.ResponseWriter, r *http.Request) {
	assignments, err := a.as.List()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = assignments
	a.IndexView.Render(w, r, vd)
}

// GET /assignments/:id
func (a *Assignments) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid assignment ID", http.StatusNotFound)
	}

	type Data struct {
		AssignmentID string
		UserID       string
		Name         string
	}
	var result Data
	a.db.Raw("SELECT a.id AS assignment_id, a.user_id, j.name FROM assignments a JOIN jobs j ON j.id=a.job_id WHERE a.id = ?", uint(id)).Scan(&result)

	// assignment, err := a.assignmentByID(w, r)
	// if err != nil {
	// 	return
	// }

	var vd views.Data
	vd.Yield = result
	a.ShowView.Render(w, r, vd)
}

// GET /assignments/:id/edit
func (a *Assignments) Edit(w http.ResponseWriter, r *http.Request) {
	assignment, err := a.assignmentByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = assignment
	a.EditView.Render(w, r, vd)
}

// POST /assignments/:id/update
func (a *Assignments) Update(w http.ResponseWriter, r *http.Request) {
	assignment, err := a.assignmentByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = assignment
	var form AssignmentForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		a.EditView.Render(w, r, vd)
		return
	}
	assignment.UserID = form.UserID

	// job, err := a.js.ByID(form.JobID)
	// if err != nil {
	// 	switch err {
	// 	case models.ErrNotFound:
	// 		http.Error(w, "Assignment not found", http.StatusNotFound)
	// 	default:
	// 		log.Println(err)
	// 		http.Error(w, "Something went terribly wrong.", http.StatusInternalServerError)
	// 	}
	// }
	// assignment.Job = *job

	err = a.as.Update(assignment)
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Assignment successfully updated!",
		}
	}
	a.EditView.Render(w, r, vd)
}

// POST /assignments
func (a *Assignments) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form AssignmentForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		a.New.Render(w, r, vd)
		return
	}

	job, err := a.js.ByID(form.JobID)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Assignment not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Something went terrible wrong.", http.StatusInternalServerError)
		}
	}

	assignment := models.Assignment{
		UserID: form.UserID,
		Job:    *job,
	}
	if err := a.as.Create(&assignment); err != nil {
		vd.SetAlert(err)
		a.New.Render(w, r, vd)
		return
	}

	url, err := a.r.Get(EditAssignment).URL("id",
		strconv.Itoa(int(assignment.ID)))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/assignments", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /assignments/:id/delete
func (a *Assignments) Delete(w http.ResponseWriter, r *http.Request) {
	assignment, err := a.assignmentByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	err = a.as.Delete(assignment.ID)
	if err != nil {
		vd.SetAlert(err)
		vd.Yield = assignment
		a.EditView.Render(w, r, vd)
		return
	}
	url, err := a.r.Get(IndexAssignments).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (a *Assignments) assignmentByID(w http.ResponseWriter, r *http.Request) (*models.Assignment, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid assignment ID", http.StatusNotFound)
		return nil, err
	}
	assignment, err := a.as.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Assignment not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Something went terribly wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}

	// job, err := a.as.Job(assignment)
	// if err != nil {
	// 	switch err {
	// 	case models.ErrNotFound:
	// 		http.Error(w, "Job (of assignemtn) not found", http.StatusNotFound)
	// 	default:
	// 		log.Println(err)
	// 		http.Error(w, "Something went terribly wrong.", http.StatusInternalServerError)
	// 	}
	// 	return nil, err
	// }
	// assignment.Job = *job

	return assignment, nil
}
