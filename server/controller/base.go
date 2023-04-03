package controller

import (
	"github.com/gin-gonic/gin"
)

var registerHandlers = make(map[string]*HandlerFunc)

func Handlers() map[string]*HandlerFunc {
	return registerHandlers
}

type HandlerFunc struct {
	Path   string
	Method string
	F      gin.HandlerFunc
}

func RegisterHandler(method, path string, f func(c *gin.Context)) {
	registerHandlers[method+path] = &HandlerFunc{
		Path:   path,
		Method: method,
		F:      f,
	}
}
