package router

import "github.com/ph0m1/p_gateway/config"

type Router interface {
	Run(cfg config.ServiceConfig)
}

type Factory interface {
	New() Router
}
