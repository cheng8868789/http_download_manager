package router

import (
	"net/http"
	"strings"

	"http_download_manager/controller"
	v1 "http_download_manager/controller/v1"

	"github.com/gin-gonic/gin"
)

func init() {
	_ = v1.Controller
}

func RegisterRouters(g *gin.Engine, basePath string) {
	var path string
	for _, v := range controller.Handlers() {
		path = basePath + v.Path
		switch strings.ToUpper(v.Method) {
		case http.MethodGet:
			g.GET(path, v.F)
		case http.MethodPut:
			g.PUT(path, v.F)
		case http.MethodPost:
			g.POST(path, v.F)
		case http.MethodDelete:
			g.DELETE(path, v.F)
		}
	}
}
