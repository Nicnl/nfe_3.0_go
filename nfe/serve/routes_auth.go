package serve

import (
	"fmt"
	jwtlib "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func (env *Env) RouteAuth(c *gin.Context) {
	var req struct {
		User string `json:"user" binding:"required"`
		Pass string `json:"pass" binding:"required"`
	}

	err := c.BindJSON(&req)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	userAdmin := true
	maxBandwidth := 0
	maxDuration := 0 * time.Second
	err = bcrypt.CompareHashAndPassword(env.AuthBlobAdmin, []byte(req.User+" / YOLO MDR PATATOTO :D / "+req.Pass))
	if err != nil {
		userAdmin = false
		maxBandwidth = 1000000 // Todo: variable d'environnement
		maxDuration = 6 * time.Hour
		err = bcrypt.CompareHashAndPassword(env.AuthBlobRegular, []byte(req.User+" / YOLO MDR PATATOTO :D / "+req.Pass))
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	// Create the token
	loginTime := time.Now().Unix()
	token := jwtlib.New(jwtlib.GetSigningMethod("HS256"))
	// Set some claims
	token.Claims = jwtlib.MapClaims{
		"login_time":    loginTime,
		"user_admin":    userAdmin,
		"max_bandwidth": maxBandwidth,
		"max_duration":  int64(maxDuration / time.Second),
	}
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(env.JwtSecret)
	if err != nil {
		c.JSON(500, gin.H{"message": "Could not generate token"})
	}

	c.JSON(200, gin.H{
		"token":         tokenString,
		"login_time":    loginTime,
		"user_admin":    userAdmin,
		"max_bandwidth": maxBandwidth,
		"max_duration":  int64(maxDuration / time.Second),
	})
}

func (env *Env) inlineAuthTransmitter(c *gin.Context) {
	authorization := c.DefaultQuery("authorization", "")
	if authorization != "" {
		if c.Request.Header.Get("Authorization") == "" {
			//fmt.Println(authorization)
			c.Request.Header.Set("Authorization", authorization)
		}
	}
}

type Claims struct {
	UserAdmin bool
}

func (env *Env) extractJwt(c *gin.Context) (*Claims, error) {
	token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(env.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwtlib.MapClaims); ok && token.Valid {
		RawUserAdmin, ok := claims["user_admin"]
		if !ok {
			return nil, fmt.Errorf("missing 'user_admin' from claims")
		}
		StrUserAdmin := fmt.Sprint(RawUserAdmin)
		if !ok {
			return nil, fmt.Errorf("'user_admin' is not a bool")
		}
		UserAdmin := StrUserAdmin == "true"

		return &Claims{
			UserAdmin: UserAdmin,
		}, nil
	} else {
		return nil, fmt.Errorf("invalid claims")
	}
}
