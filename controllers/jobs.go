package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sirodoht/heartfort/models"
	"github.com/sirodoht/heartfort/views"
)

const (
	IndexJobs = "index_jobs"
	ShowJob   = "show_job"
	EditJob   = "edit_job"

	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewJobs(js models.JobService, r *mux.Router) *Jobs {
	return &Jobs{
		New:       views.NewView("layout", "jobs/new"),
		ShowView:  views.NewView("layout", "jobs/show"),
		EditView:  views.NewView("layout", "jobs/edit"),
		IndexView: views.NewView("layout", "jobs/index"),
		js:        js,
		r:         r,
	}
}

type Jobs struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	js        models.JobService
	r         *mux.Router
}

type JobForm struct {
	Name string `schema:"name"`
}

// GET /jobs
func (j *Jobs) Index(w http.ResponseWriter, r *http.Request) {
	jobs, err := j.js.List()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = jobs
	j.IndexView.Render(w, r, vd)
}

// GET /jobs/:id
func (j *Jobs) Show(w http.ResponseWriter, r *http.Request) {
	job, err := j.jobByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = job
	j.ShowView.Render(w, r, vd)
}

// GET /jobs/:id/edit
func (j *Jobs) Edit(w http.ResponseWriter, r *http.Request) {
	job, err := j.jobByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = job
	j.EditView.Render(w, r, vd)
}

// POST /jobs/:id/update
func (j *Jobs) Update(w http.ResponseWriter, r *http.Request) {
	job, err := j.jobByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = job
	var form JobForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		j.EditView.Render(w, r, vd)
		return
	}
	job.Name = form.Name
	err = j.js.Update(job)
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Job successfully updated!",
		}
	}
	j.EditView.Render(w, r, vd)
}

// POST /jobs
func (j *Jobs) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form JobForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		j.New.Render(w, r, vd)
		return
	}
	job := models.Job{
		Name: form.Name,
	}
	if err := j.js.Create(&job); err != nil {
		vd.SetAlert(err)
		j.New.Render(w, r, vd)
		return
	}

	url, err := j.r.Get(EditJob).URL("id",
		strconv.Itoa(int(job.ID)))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/jobs", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /jobs/:id/delete
func (j *Jobs) Delete(w http.ResponseWriter, r *http.Request) {
	job, err := j.jobByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	err = j.js.Delete(job.ID)
	if err != nil {
		vd.SetAlert(err)
		vd.Yield = job
		j.EditView.Render(w, r, vd)
		return
	}
	url, err := j.r.Get(IndexJobs).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (j *Jobs) jobByID(w http.ResponseWriter,
	r *http.Request) (*models.Job, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid job ID", http.StatusNotFound)
		return nil, err
	}
	job, err := j.js.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Job not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Something terrible wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	return job, nil
}
