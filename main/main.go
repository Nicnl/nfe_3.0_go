package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

	authBlobRegular, err := bcrypt.GenerateFromPassword([]byte("user"+" / YOLO MDR PATATOTO :D / "+"user"), 11)
	if err != nil {
		panic(err)
	}

	authBlobAdmin, err := bcrypt.GenerateFromPassword([]byte("admin"+" / YOLO MDR PATATOTO :D / "+"admin"), 11)
	if err != nil {
		panic(err)
	}

	env := serve.Env{
		Vfs:             vfs.New("/vmshare_hub/ISOs/"),
		Transfers:       map[string]*transfer.Transfer{},
		AuthBlobRegular: authBlobRegular,
		AuthBlobAdmin:   authBlobAdmin,
	}

	routerApi.POST("/auth", env.RouteAuth)

	routerApi.GET("/guest/:basepath/ls", env.RouteGuestLsRoot)
	routerApi.GET("/guest/:basepath/ls/", env.RouteGuestLsRoot)
	routerApi.GET("/guest/:basepath/ls/:path", env.RouteGuestLs)
	routerApi.GET("/guest/:basepath/ls/:path/", env.RouteGuestLs)
	routerApiPrivate := routerApi.Group("/")
	routerApiPrivate.Use(jwt.Auth(string(env.JwtSecret)))
	{
		routerApiPrivate.GET("/transfers", env.RouteTransfersList)
		routerApiPrivate.DELETE("/transfers", env.RouteTransfersClear)
		routerApiPrivate.DELETE("/transfer/:guid/", env.RouteTransferInterrupt)
		routerApiPrivate.PATCH("/transfer/:guid/", env.RouteTransferChangeSpeed)

		routerApiPrivate.GET("/ls/", env.RouteAuthLsRoot)
		routerApiPrivate.GET("/ls/:path", env.RouteAuthLs)

		routerApiPrivate.POST("/gen/", env.RouteAuthRegenLink)
	}

	routerDownload.GET("/:path", env.RouteDownload)
	routerDownload.GET("/:path/:osef", env.RouteDownload)

	routerDownload.GET("/:path/:osef/guest/:realpath", env.RouteDownloadGuest)
	routerDownload.GET("/:path/:osef/guest/:realpath/:osef2", env.RouteDownloadGuest)

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
