package serve

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/transfer"
)

func (env *Env) RouteTransfersList(c *gin.Context) {
	claims, err := env.extractJwt(c)
	if err != nil {
		panic(err)
	}

	if !claims.UserAdmin {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	c.JSON(http.StatusOK, env.TransfersCopy())
}

func (env *Env) RouteTransfersClear(c *gin.Context) {
	claims, err := env.extractJwt(c)
	if err != nil {
		panic(err)
	}

	if !claims.UserAdmin {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	env.TransfersDelete(func(key string, val *transfer.Transfer) bool {
		return val.CurrentState != transfer.StateTransferring
	})

	c.Status(http.StatusNoContent)
}

func (env *Env) RouteTransferInterrupt(c *gin.Context) {
	claims, err := env.extractJwt(c)
	if err != nil {
		panic(err)
	}

	if !claims.UserAdmin {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	guid := c.Param("guid")

	t, ok := env.TransfersGet(guid)
	if !ok {
		c.String(http.StatusNotFound, "unknown transfer guid")
		return
	}

	t.ShouldInterrupt = true
	c.Status(http.StatusNoContent)
}

func (env *Env) RouteTransferChangeSpeed(c *gin.Context) {
	claims, err := env.extractJwt(c)
	if err != nil {
		panic(err)
	}

	if !claims.UserAdmin {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	guid := c.Param("guid")

	var request struct {
		SpeedLimit int64 `json:"speed_limit"`
	}

	err = c.BindJSON(&request)
	if err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	t, ok := env.TransfersGet(guid)
	if !ok {
		c.String(http.StatusNotFound, "unknown transfer guid")
		return
	}

	var bufferSize int64 = 50 * 1024

	if request.SpeedLimit > 0 && bufferSize*4 > request.SpeedLimit {
		if request.SpeedLimit >= 4 {
			bufferSize = request.SpeedLimit / 4
		} else {
			bufferSize = 1
		}
	}

	t.ChangeBufferSize(bufferSize)
	t.SetSpeedLimit(request.SpeedLimit)
	c.Status(http.StatusNoContent)
}
