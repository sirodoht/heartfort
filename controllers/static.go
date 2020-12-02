package controllers

import "github.com/sirodoht/heartfort/views"

func NewStatic() *Static {
	return &Static{
		Home:  views.NewView("layout", "static/home"),
		Specs: views.NewView("layout", "static/specs"),
	}
}

type Static struct {
	Home  *views.View
	Specs *views.View
}
