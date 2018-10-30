package serve

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"nfe_3.0_go/nfe/transfer"
	"time"
)

func (env *Env) RouteDownloadGuest(c *gin.Context) {
	basePath, speedLimit, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link 1")
		return
	}

	key := crypt.GlobUnique([]byte(crypt.PathEncode(c.Param("path"))))

	subVfs := env.Vfs.SubVfs(basePath)

	vfsPath, _, err := crypt.Find(crypt.HexDecode(c.Param("realpath"), key), time.Now(), subVfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link 2")
		panic(err)
		return
	}

	var bufferSize int64 = 50 * 1024

	if speedLimit > 0 && bufferSize*4 > speedLimit {
		if speedLimit >= 4 {
			bufferSize = speedLimit / 4
		} else {
			bufferSize = 1
		}
	}

	// Initialisation d'un nouveau transfert
	t, err := transfer.New(subVfs, vfsPath, speedLimit, bufferSize)
	if err != nil {
		panic(err)
	}

	// Enregistrement du transfert dans la liste globale
	env.TransfersSet(t.Guid.String(), t)
	//defer delete(transfers, t.Guid.String())

	// Envoi du fichier
	env.ServeFile(c, t, subVfs)
}

func (env *Env) RouteDownload(c *gin.Context) {
	vfsPath, speedLimit, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid link")
		return
	}

	var bufferSize int64 = 50 * 1024

	if speedLimit > 0 && bufferSize*4 > speedLimit {
		if speedLimit >= 4 {
			bufferSize = speedLimit / 4
		} else {
			bufferSize = 1
		}
	}

	// Initialisation d'un nouveau transfert
	t, err := transfer.New(env.Vfs, vfsPath, speedLimit, bufferSize)
	if err != nil {
		panic(err)
	}

	// Enregistrement du transfert dans la liste globale
	env.TransfersSet(t.Guid.String(), t)
	//defer delete(transfers, t.Guid.String())

	// Envoi du fichier
	env.ServeFile(c, t, env.Vfs)
}
