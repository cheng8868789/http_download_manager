package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"http_download_manager/core"
	"http_download_manager/middle"
	"http_download_manager/router"
)

func main() {

	gin.SetMode(gin.DebugMode)
	engine := gin.New()
	engine.Use(middle.Core())
	engine.Use(gin.Recovery())
	v1 := engine.Group("/api")
	router.RegisterRouters(engine, v1.BasePath())

	err := start()
	if err != nil {
		panic(err)
	}

	err = engine.Run(fmt.Sprintf(":%d", 8081))
	if err != nil {
		panic(err)
	}
}

func start() error {
	err := core.Init()
	if err != nil {
		return err
	}
	return nil
}
