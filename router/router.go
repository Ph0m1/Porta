package router

import "github.com/ph0m1/porta/config"

type Router interface {
	Run(cfg config.ServiceConfig)
}

type Factory interface {
	New() Router
}
