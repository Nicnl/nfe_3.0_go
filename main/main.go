package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/transfer"
	"nfe_3.0_go/nfe/vfs"
	"os"
	"time"
)

func startRouter(channel chan error, addr string, handler http.Handler) {
	channel <- http.ListenAndServe(addr, handler)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	//path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	routerDownload := gin.Default()
	routerApi := gin.Default()

	routerDownload.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "HEAD", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	}))

	routerApi.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "HEAD", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	}))

	env := serve.Env{
		Vfs:       vfs.New("/vmshare_hub/ISOs/"),
		Transfers: map[string]*transfer.Transfer{},
	}

	routerApi.GET("/transfers", env.RouteTransfersList)
	routerApi.DELETE("/transfer/:guid/", env.RouteTransferInterrupt)
	routerApi.PATCH("/transfer/:guid/", env.RouteTransferChangeSpeed)

	routerApi.GET("/ls/", env.RouteAuthLsRoot)
	routerApi.GET("/ls/:path", env.RouteAuthLs)

	routerApi.POST("/gen/", env.RouteAuthRegenLink)

	routerDownload.GET("/:path", env.RouteDownload)
	routerDownload.GET("/:path/:osef", env.RouteDownload)

	channel := make(chan error)
	go startRouter(channel, ":9000", routerDownload)
	go startRouter(channel, ":9001", routerApi)

	if err := <-channel; err != nil {
		panic(err)
	} else {
		fmt.Println("An http handler has stopped without throwing an error.")
	}
	os.Exit(1)
}
