package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"math/rand"
	"net/http"
	"nfe_3.0_go/in_memory_content_serving"
	"nfe_3.0_go/nfe/crypt"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/vfs"
	"os"
	"strconv"
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

	var err error

	var nonAdminSpeedLimit int64 = 0
	if v := os.Getenv("NON_ADMIN_SPEED_LIMIT"); v != "" {
		nonAdminSpeedLimit, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Println("Error while parsing NON_ADMIN_SPEED_LIMIT as an int64")
		}
	}

	var nonAdminTimeLimit int64 = 6 * 60 * 60
	if v := os.Getenv("NON_ADMIN_TIME_LIMIT"); v != "" {
		nonAdminTimeLimit, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Println("Error while parsing NON_ADMIN_TIME_LIMIT as an int64")
		}
	}

	var defaultSpeedLimit int64 = 0
	if v := os.Getenv("DEFAULT_SPEED_LIMIT"); v != "" {
		defaultSpeedLimit, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Println("Error while parsing DEFAULT_SPEED_LIMIT as an int64")
		}
	}

	var defaultTimeLimit int64 = 15 * 60
	if v := os.Getenv("DEFAULT_TIME_LIMIT"); v != "" {
		defaultTimeLimit, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Println("Error while parsing DEFAULT_TIME_LIMIT as an int64")
		}
	}

	env := serve.Env{
		Vfs: vfs.New(os.Getenv("BASE_PATH")),
		//Transfers: map[string]*transfer.Transfer{},

		BasePath:           os.Getenv("BASE_PATH"),
		PasswordHashSalt:   []byte("YOLO MDR PATATOTOOTOOOOOO :DDQSDPOIQSDOIUQS NFE NFE NFE NFE 3.0 YOUPI LOL HEHE YOY MDR YOOOY DOIUQSD #{#[[|\\`^`|[{#|`\\=))à)à`" + os.Getenv("JWT_SECRET")),
		JwtSecret:          []byte(os.Getenv("JWT_SECRET")),
		AuthBlobRegular:    []byte(os.Getenv("PW_HASH_USER")),
		AuthBlobAdmin:      []byte(os.Getenv("PW_HASH_ADMIN")),
		NonAdminSpeedLimit: nonAdminSpeedLimit,
		NonAdminTimeLimit:  nonAdminTimeLimit,
		DefaultSpeedLimit:  defaultSpeedLimit,
		DefaultTimeLimit:   defaultTimeLimit,
		GlobSalt:           []byte(os.Getenv("GLOB_SALT_LEGACY")),
		GlobUrlList:        []byte(os.Getenv("GLOB_SALT_LIST")),
		GlobUrlDown:        []byte(os.Getenv("GLOB_SALT_DOWN")),
	}

	env.TransfersInit()

	crypt.GlobSalt = env.GlobSalt
	crypt.GlobUrlList = env.GlobUrlList
	crypt.GlobUrlDown = env.GlobUrlDown

	routerApi.POST("/api/auth", env.RouteAuth)
	routerApi.POST("/api/hash", env.RequestHash)
	routerApi.GET("/api/is_configured", env.CheckAuthConfigured)

	routerApi.GET("/api/guest/:basepath/ls", env.RouteGuestLsRoot)
	routerApi.GET("/api/guest/:basepath/ls/", env.RouteGuestLsRoot)
	routerApi.GET("/api/guest/:basepath/ls/:path", env.RouteGuestLs)
	routerApi.GET("/api/guest/:basepath/ls/:path/", env.RouteGuestLs)
	routerApiPrivate := routerApi.Group("/")
	routerApiPrivate.Use(jwt.Auth(string(env.JwtSecret)))
	{
		routerApiPrivate.GET("/api/transfers", env.RouteTransfersList)
		routerApiPrivate.DELETE("/api/transfers", env.RouteTransfersClear)
		routerApiPrivate.DELETE("/api/transfer/:guid/", env.RouteTransferInterrupt)
		routerApiPrivate.PATCH("/api/transfer/:guid/", env.RouteTransferChangeSpeed)

		routerApiPrivate.GET("/api/ls/", env.RouteAuthLsRoot)
		routerApiPrivate.GET("/api/ls/:path", env.RouteAuthLs)

		routerApiPrivate.POST("/api/gen/", env.RouteAuthRegenLink)
	}

	err = in_memory_content_serving.PopulateRouter(routerApi, "web", 15*60)
	if err != nil {
		fmt.Println("Error when populating routerApi with in-memory-content-serving website data")
		panic(err)
	}

	routerApi.GET("/static/config.js", func(c *gin.Context) {
		var config struct {
			UrlApi  string `json:"urlApi"`
			UrlDown string `json:"urlDown"`
		}

		config.UrlApi = os.Getenv("URL_LIST")
		config.UrlDown = os.Getenv("URL_DOWN")

		data, err := json.Marshal(&config)
		if err != nil {
			panic(err)
		}
		fmt.Fprint(c.Writer, "window.appConfig=", string(data), ";")
	})

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
