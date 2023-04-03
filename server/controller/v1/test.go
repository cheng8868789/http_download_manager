package v1

import (
	"fmt"
	"http_download_manager/controller"
	"http_download_manager/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

var Controller int

func init() {
	controller.RegisterHandler(http.MethodGet, "/v1/httpDownLoad", httpDownLoad)
	controller.RegisterHandler(http.MethodGet, "/v1/downLoadProcess", downLoadProcess)
	controller.RegisterHandler(http.MethodGet, "/v1/stopDownLoad", stopDownLoad)
}

// /api/v1/httpDownLoad?url=
func httpDownLoad(c *gin.Context) {
	fmt.Println("get request")
	url := c.Query("url")
	go core.DownLoad(url)
	fmt.Println("\nDownloading")
	c.JSON(http.StatusOK, "Downloading")
}

func downLoadProcess(c *gin.Context) {
	fileList := core.DownLoadProcess()
	c.JSON(http.StatusOK, fileList)

}

func stopDownLoad(c *gin.Context) {
	fileName := c.Query("fileName")
	err := core.StopDownLoad(fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		c.JSON(http.StatusOK, "文件删除成功")
	}
}
