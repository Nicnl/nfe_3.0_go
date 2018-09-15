package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/transfer"
	"nfe_3.0_go/nfe/vfs"
	"os"
	"path/filepath"
	"time"
)

func startRouter(channel chan error, addr string, handler http.Handler) {
	channel <- http.ListenAndServe(addr, handler)
}

type JsonDir struct {
	Name    string `json:"name"`
	VfsPath string `json:"path"`
}

type JsonFile struct {
	Name    string `json:"name"`
	VfsPath string `json:"path"`
	Size    int64  `json:"size"`
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

	transfers := map[string]*transfer.Transfer{}

	env := serve.Env{
		Vfs: vfs.New("/vmshare_hub/ISOs/"),
	}

	routerApi.GET("/transfers", func(c *gin.Context) {
		c.JSON(http.StatusOK, transfers)
	})

	routerApi.DELETE("/transfer/:guid/", func(c *gin.Context) {
		guid := c.Param("guid")

		t, ok := transfers[guid]
		if !ok {
			c.String(http.StatusNotFound, "unknown transfer guid")
			return
		}

		t.ShouldInterrupt = true
		c.Status(http.StatusNoContent)
	})

	routerApi.PATCH("/transfer/:guid/", func(c *gin.Context) { // Todo: en faire un patch avec les données en JSON
		guid := c.Param("guid")

		var request struct {
			SpeedLimit int64 `json:"speed_limit"`
		}

		err := c.BindJSON(&request)
		if err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}

		t, ok := transfers[guid]
		if !ok {
			c.String(http.StatusNotFound, "unknown transfer guid")
			return
		}

		t.SetSpeedLimit(request.SpeedLimit)
		c.Status(http.StatusNoContent)
	})

	routerApi.GET("/ls/", func(c *gin.Context) {
		rawFiles, err := env.Vfs.Ls("/")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		var output struct {
			Path       string     `json:"path"`
			ParentPath *string    `json:"parent_path"`
			Dirs       []JsonDir  `json:"dirs"`
			Files      []JsonFile `json:"files"`
		}

		output.Path = "/"
		output.ParentPath = nil
		output.Dirs = []JsonDir{}
		output.Files = []JsonFile{}

		for _, rawFile := range rawFiles {
			vfsPath := "/" + rawFile

			info, err := env.Vfs.Stat(vfsPath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				fmt.Fprintln(c.Writer, err)
				return
			}

			if info.IsDir() {
				output.Dirs = append(output.Dirs, JsonDir{
					Name:    rawFile,
					VfsPath: crypt.PathEncode(vfsPath),
				})
			} else {
				output.Files = append(output.Files, JsonFile{
					Name:    rawFile,
					VfsPath: crypt.PathEncodeExpirable(vfsPath, 1*time.Minute, time.Now()), // Todo : temps par défaut configurable
					Size:    info.Size(),
				})

				// Todo: ajouter limite de débit par défaut
			}
		}

		c.JSON(http.StatusOK, &output)
	})

	routerApi.GET("/ls/:path", func(c *gin.Context) {
		path, _, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
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
			Path       string     `json:"path"`
			ParentPath *string    `json:"parent_path"`
			Dirs       []JsonDir  `json:"dirs"`
			Files      []JsonFile `json:"files"`
		}

		output.Path = path
		parentPath := filepath.Dir(path)
		if parentPath == "/" {
			parentPath = ""
		} else {
			parentPath = crypt.PathEncode(parentPath)
		}
		output.ParentPath = &parentPath
		output.Dirs = []JsonDir{}
		output.Files = []JsonFile{}

		for _, rawFile := range rawFiles {
			vfsPath := path + "/" + rawFile

			info, err := env.Vfs.Stat(vfsPath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				fmt.Fprintln(c.Writer, err)
				return
			}

			if info.IsDir() {
				output.Dirs = append(output.Dirs, JsonDir{
					Name:    rawFile,
					VfsPath: crypt.PathEncode(vfsPath),
				})
			} else {
				output.Files = append(output.Files, JsonFile{
					Name:    rawFile,
					VfsPath: crypt.PathEncodeExpirable(vfsPath, 1*time.Minute, time.Now()), // Todo : temps par défaut configurable
					Size:    info.Size(),
				})

				// Todo: ajouter limite de débit par défaut
			}
		}

		c.JSON(http.StatusOK, &output)
	})

	routerApi.GET("/gen/:path", func(c *gin.Context) {
		path, _, err := crypt.FindTimeLimitIgnorable(c.Param("path"), time.Now(), env.Vfs, true)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		var out struct {
			Path string `json:"path"`
		}

		out.Path = crypt.PathEncodeExpirable(path, 30*time.Minute, time.Now()) // Todo: déplacer paramètre de durée vers ?GET + ajouter limite de bande passante

		c.JSON(http.StatusOK, &out)
	})

	routerDownload.GET("/:path", func(c *gin.Context) {
		vfsPath, _, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
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
		//defer delete(transfers, t.Guid.String())

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