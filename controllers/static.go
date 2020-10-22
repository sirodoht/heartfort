package controllers

import "github.com/sirodoht/heartfort/views"

func NewStatic() *Static {
	return &Static{
		Home: views.NewView("layout", "static/home"),
	}
}

type Static struct {
	Home *views.View
}
