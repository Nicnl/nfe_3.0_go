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

func (env *Env) RouteGuestLsRoot(c *gin.Context) {
	basePath, _, err := crypt.Find(c.Param("basepath"), time.Now(), env.Vfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link hehe")
		panic(err)
		return
	}

	key := crypt.GlobUnique([]byte(crypt.PathEncode(c.Param("basepath"))))

	env.routeLs(c, basePath, "/", key)
}

func (env *Env) RouteGuestLs(c *gin.Context) {
	basePath, _, err := crypt.Find(c.Param("basepath"), time.Now(), env.Vfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link")
		return
	}

	key := crypt.GlobUnique([]byte(crypt.PathEncode(c.Param("basepath"))))

	path, _, err := crypt.Find(crypt.HexDecode(c.Param("path"), key), time.Now(), env.Vfs.SubVfs(basePath))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link")
		return
	}

	env.routeLs(c, basePath, path, key)
}

func (env *Env) RouteAuthLsRoot(c *gin.Context) {
	env.routeLs(c, "", "/", "")
}

func (env *Env) RouteAuthLs(c *gin.Context) {
	path, _, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link")
		return
	}

	env.routeLs(c, "", path, "")
}

func (env *Env) routeLs(c *gin.Context, basePath string, path string, key string) {
	subVfs := env.Vfs.SubVfs(basePath)
	rawFiles, err := subVfs.Ls(path)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link")
		panic(err)
		return
	}

	var output struct {
		Path       string     `json:"path"`
		ParentPath *string    `json:"parent_path"`
		Dirs       []jsonDir  `json:"dirs"`
		Files      []jsonFile `json:"files"`
	}

	output.Path = path

	if path != "/" {
		parentPath := filepath.Dir(path)
		if parentPath == "/" {
			parentPath = ""
		} else {
			parentPath = crypt.PathEncode(parentPath)
			if key != "" {
				parentPath = crypt.HexEncode(parentPath, key)
			}
		}
		output.ParentPath = &parentPath
	}

	output.Dirs = []jsonDir{}
	output.Files = []jsonFile{}

	for _, rawFile := range rawFiles {
		vfsPath := path + "/" + rawFile

		info, err := subVfs.Stat(vfsPath)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Fprintln(c.Writer, err)
			return
		}

		if info.IsDir() {
			encodedPath := crypt.PathEncode(vfsPath)
			if key != "" {
				encodedPath = crypt.HexEncode(encodedPath, key)
			}

			output.Dirs = append(output.Dirs, jsonDir{
				Name:    rawFile,
				VfsPath: encodedPath,
			})
		} else {
			var encodedPath string
			if key != "" {
				encodedPath = crypt.HexEncode(crypt.PathEncode(vfsPath), key)
			} else {
				encodedPath = crypt.PathEncodeExpirable(vfsPath, 1*time.Minute, time.Now())
			}

			output.Files = append(output.Files, jsonFile{
				Name:    rawFile,
				VfsPath: encodedPath, // Todo : temps par défaut configurable
				Size:    info.Size(),
			})

			// Todo: ajouter limite de débit par défaut
		}
	}

	c.JSON(http.StatusOK, &output)
}
