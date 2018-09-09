package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/serve"
	"nfe_3.0_go/nfe/transfer"
	"strconv"
)

func main() {
	path := "/vmshare_hub/ISOs/Windows/Windows 10 - 1703/Win10_1703_French_x64.iso"

	r := gin.Default()

	transfers := map[string]*transfer.Transfer{}

	r.GET("/download", func(c *gin.Context) {
		// Initialisation d'un nouveau transfert
		t, err := transfer.New(path, 0, 20*1024)
		if err != nil {
			panic(err)
		}

		// Enregistrement du transfert dans la liste globale
		transfers[t.Guid.String()] = t
		defer delete(transfers, t.Guid.String())

		// Envoi du fichier
		serve.ServeFile(c, t)
	})

	r.GET("/transfers", func(c *gin.Context) {
		c.JSON(http.StatusOK, transfers)
	})

	r.GET("/transfer/:guid/set_speed_limit/:speed_limit", func(c *gin.Context) {
		guid := c.Param("guid")

		speedLimit, err := strconv.ParseInt(c.Param("speed_limit"), 10, 64)
		if err != nil {
			panic(err)
		}

		t, ok := transfers[guid]
		if !ok {
			panic(fmt.Errorf("unknown transfer with guid '%d'", guid))
		}

		t.SetSpeedLimit(speedLimit)
		c.Status(http.StatusOK)
		fmt.Fprintln(c.Writer, "OK")
	})

	err := http.ListenAndServe(":9000", r)
	if err != nil {
		panic(err)
	}
}
