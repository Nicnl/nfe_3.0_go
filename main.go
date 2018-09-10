package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/transfer"
	"nfe_3.0_go/nfe/vfs"
	"os"
	"strconv"
	"time"
)

func startRouter(channel chan error, addr string, handler http.Handler) {
	channel <- http.ListenAndServe(addr, handler)
}

func main() {
	//path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	routerDownload := gin.Default()
	routerApi := gin.Default()

	transfers := map[string]*transfer.Transfer{}

	env := serve.Env{
		Vfs: vfs.New("/vmshare_hub/ISOs/"),
	}

	routerApi.GET("/transfers", func(c *gin.Context) {
		c.JSON(http.StatusOK, transfers)
	})

	routerApi.GET("/transfer/:guid/set_speed_limit/:speed_limit", func(c *gin.Context) {
		guid := c.Param("guid")

		speedLimit, err := strconv.ParseInt(c.Param("speed_limit"), 10, 64)
		if err != nil {
			panic(err)
		}

		t, ok := transfers[guid]
		if !ok {
			panic(fmt.Errorf("unknown transfer with guid '%s'", guid))
		}

		t.SetSpeedLimit(speedLimit)
		c.Status(http.StatusOK)
		fmt.Fprintln(c.Writer, "OK")
	})

	routerApi.GET("/ls/", func(c *gin.Context) {
		rawFiles, err := env.Vfs.Ls("/")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		var output struct {
			Path  string            `json:"path"`
			Dirs  map[string]string `json:"dirs"`
			Files map[string]string `json:"files"`
		}

		output.Dirs = make(map[string]string)
		output.Files = make(map[string]string)

		output.Path = "/"
		for _, rawFile := range rawFiles {
			vfsPath := "/" + rawFile

			info, err := env.Vfs.Stat(vfsPath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				fmt.Fprintln(c.Writer, err)
				return
			}

			if info.IsDir() {
				output.Dirs[rawFile] = crypt.PathEncode(vfsPath)
			} else {
				output.Files[rawFile] = crypt.PathEncodeExpirable(vfsPath, time.Second*30, time.Now())
			}
		}

		c.JSON(http.StatusOK, &output)
	})

	routerApi.GET("/ls/:path", func(c *gin.Context) {
		path, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		rawFiles, err := env.Vfs.Ls(path)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		var output struct {
			Path  string            `json:"path"`
			Dirs  map[string]string `json:"dirs"`
			Files map[string]string `json:"files"`
		}

		output.Dirs = make(map[string]string)
		output.Files = make(map[string]string)

		output.Path = path
		for _, rawFile := range rawFiles {
			vfsPath := path + "/" + rawFile

			info, err := env.Vfs.Stat(vfsPath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				fmt.Fprintln(c.Writer, err)
				return
			}

			if info.IsDir() {
				output.Dirs[rawFile] = crypt.PathEncode(vfsPath)
			} else {
				output.Files[rawFile] = crypt.PathEncodeExpirable(vfsPath, time.Second*30, time.Now())
			}
		}

		c.JSON(http.StatusOK, &output)
	})

	routerDownload.GET("/:path", func(c *gin.Context) {
		vfsPath, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		// Initialisation d'un nouveau transfert
		t, err := transfer.New(env.Vfs, vfsPath, 0, 20*1024)
		if err != nil {
			panic(err)
		}

		// Enregistrement du transfert dans la liste globale
		transfers[t.Guid.String()] = t
		defer delete(transfers, t.Guid.String())

		// Envoi du fichier
		env.ServeFile(c, t)
	})

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
