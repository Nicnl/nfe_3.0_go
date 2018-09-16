package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"nfe_3.0_go/nfe/transfer"
	"time"
)

func (env *Env) RouteDownload(c *gin.Context) {
	vfsPath, speedLimit, err := crypt.Find(c.Param("path"), time.Now(), env.Vfs)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Fprintln(c.Writer, err)
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
	env.Transfers[t.Guid.String()] = t
	//defer delete(transfers, t.Guid.String())

	// Envoi du fichier
	env.ServeFile(c, t)
}
