package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"path/filepath"
	"time"
)

type jsonDir struct {
	Name    string `json:"name"`
	VfsPath string `json:"path"`
}

type jsonFile struct {
	Name    string `json:"name"`
	VfsPath string `json:"path"`
	Size    int64  `json:"size"`
}

func (env *Env) RouteAuthLsRoot(c *gin.Context) {
	env.routeAuthLs(c, "/")
}

func (env *Env) RouteAuthLs(c *gin.Context) {
	path, _, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Fprintln(c.Writer, err)
		return
	}

	env.routeAuthLs(c, path)
}

func (env *Env) routeAuthLs(c *gin.Context, path string) {
	rawFiles, err := env.Vfs.Ls(path)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Fprintln(c.Writer, err)
		return
	}

	var output struct {
		Path       string     `json:"path"`
		ParentPath *string    `json:"parent_path"`
		Dirs       []jsonDir  `json:"dirs"`
		Files      []jsonFile `json:"files"`
	}

	output.Path = path
	parentPath := filepath.Dir(path)
	if parentPath == "/" {
		parentPath = ""
	} else {
		parentPath = crypt.PathEncode(parentPath)
	}
	output.ParentPath = &parentPath
	output.Dirs = []jsonDir{}
	output.Files = []jsonFile{}

	for _, rawFile := range rawFiles {
		vfsPath := path + "/" + rawFile

		info, err := env.Vfs.Stat(vfsPath)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		if info.IsDir() {
			output.Dirs = append(output.Dirs, jsonDir{
				Name:    rawFile,
				VfsPath: crypt.PathEncode(vfsPath),
			})
		} else {
			output.Files = append(output.Files, jsonFile{
				Name:    rawFile,
				VfsPath: crypt.PathEncodeExpirable(vfsPath, 1*time.Minute, time.Now()), // Todo : temps par défaut configurable
				Size:    info.Size(),
			})

			// Todo: ajouter limite de débit par défaut
		}
	}

	c.JSON(http.StatusOK, &output)
}
