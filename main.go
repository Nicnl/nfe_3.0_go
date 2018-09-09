package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/transfer"
)

func main() {
	path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	r := gin.Default()

	r.GET("/download", func(c *gin.Context) {
		t, err := transfer.New(path, 0, 20*1024)
		if err != nil {
			panic(err)
		}

		serve.ServeFile(c, t)
	})

	err := http.ListenAndServe(":9000", r)
	if err != nil {
		panic(err)
	}
}
