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
	IndexMates = "index_mates"
	ShowMate   = "show_mate"
	EditMate   = "edit_mate"
)

func NewMates(ms models.MateService, r *mux.Router) *Mates {
	return &Mates{
		New:       views.NewView("layout", "mates/new"),
		ShowView:  views.NewView("layout", "mates/show"),
		EditView:  views.NewView("layout", "mates/edit"),
		IndexView: views.NewView("layout", "mates/index"),
		ms:        ms,
		r:         r,
	}
}

type Mates struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	ms        models.MateService
	r         *mux.Router
}

type MateForm struct {
	Email string `schema:"email"`
}

// GET /mates
func (m *Mates) Index(w http.ResponseWriter, r *http.Request) {
	mates, err := m.ms.List()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = mates
	m.IndexView.Render(w, r, vd)
}

// GET /mates/:id
func (m *Mates) Show(w http.ResponseWriter, r *http.Request) {
	mate, err := m.mateById(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = mate
	m.ShowView.Render(w, r, vd)
}

// GET /mates/:id/edit
func (m *Mates) Edit(w http.ResponseWriter, r *http.Request) {
	mate, err := m.mateById(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = mate
	m.EditView.Render(w, r, vd)
}

// POST /mates/:id/update
func (m *Mates) Update(w http.ResponseWriter, r *http.Request) {
	mate, err := m.mateById(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yield = mate
	var form MateForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		m.EditView.Render(w, r, vd)
		return
	}
	mate.Email = form.Email
	err = m.ms.Update(mate)
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Mate successfully updated!",
		}
	}
	m.EditView.Render(w, r, vd)
}

// POST /mates
func (m *Mates) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form MateForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		m.New.Render(w, r, vd)
		return
	}
	mate := models.Mate{
		Email: form.Email,
	}
	if err := m.ms.Create(&mate); err != nil {
		vd.SetAlert(err)
		m.New.Render(w, r, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Email registered.",
	}

	m.New.Render(w, r, vd)
}

// POST /mates/:id/delete
func (m *Mates) Delete(w http.ResponseWriter, r *http.Request) {
	mate, err := m.mateById(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	err = m.ms.Delete(mate.ID)
	if err != nil {
		vd.SetAlert(err)
		vd.Yield = mate
		m.EditView.Render(w, r, vd)
		return
	}
	url, err := m.r.Get(IndexMates).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (m *Mates) mateById(w http.ResponseWriter, r *http.Request) (*models.Mate, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid mate ID", http.StatusNotFound)
		return nil, err
	}
	mate, err := m.ms.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Mate not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Something terrible wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	return mate, nil
}
