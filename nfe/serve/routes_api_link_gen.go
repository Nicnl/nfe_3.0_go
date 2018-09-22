package serve

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"nfe_3.0_go/nfe/crypt"
	"time"
)

func (env *Env) RouteAuthRegenLink(c *gin.Context) {
	claims, err := env.extractJwt(c)
	if err != nil {
		panic(err)
	}

	var request struct {
		Path     string `json:"path"`
		Duration int64  `json:"duration"`
		Speed    int64  `json:"speed"`
	}

	err = c.BindJSON(&request)
	if err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	if !claims.UserAdmin {
		if request.Duration > env.NonAdminTimeLimit {
			request.Duration = env.NonAdminTimeLimit
		}
		if request.Speed > env.NonAdminSpeedLimit {
			request.Speed = env.NonAdminSpeedLimit
		}
	}

	path, _, err := crypt.FindTimeLimitIgnorable(request.Path, time.Now(), env.Vfs, true)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Fprintln(c.Writer, err)
		return
	}

	var out struct {
		Path string `json:"path"`
	}

	if request.Duration > 0 {
		out.Path = crypt.PathEncodeExpirable(path, time.Duration(request.Duration)*time.Second, time.Now())
	} else {
		out.Path = crypt.PathEncode(path)
	}

	if request.Speed > 0 {
		out.Path = crypt.AddBandwidthLimit(out.Path, request.Speed)
	}

	c.JSON(http.StatusOK, &out)
}
