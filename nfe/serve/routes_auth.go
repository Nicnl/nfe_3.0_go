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

func (env *Env) CheckAuthConfigured(c *gin.Context) {
	var response struct {
		IsConfigured bool     `json:"is_configured"`
		Message      []string `json:"messages"`
	}

	response.IsConfigured = true
	response.Message = make([]string, 0)

	if len(env.JwtSecret) == 0 {
		response.Message = append(response.Message, "La variable d'environnement JWT_SECRET est vide : doit contenir une secret pour l'encryption. (remplir au pif)")
		response.IsConfigured = false
	}

	if len(env.BasePath) == 0 {
		response.Message = append(response.Message, "La variable d'environnement BASE_PATH est vide : doit contenir le chemin du partage de fichiers.")
		response.IsConfigured = false
	}

	if len(env.GlobSalt) == 0 {
		response.Message = append(response.Message, "La variable d'environnement GLOB_SALT_LEGACY est vide : doit contenir une secret pour l'encryption. (remplir au pif)")
		response.IsConfigured = false
	}

	if len(env.GlobUrlList) == 0 {
		response.Message = append(response.Message, "La variable d'environnement GLOB_SALT_LIST est vide : doit contenir une secret pour l'encryption. (remplir au pif)")
		response.IsConfigured = false
	}

	if len(env.GlobUrlDown) == 0 {
		response.Message = append(response.Message, "La variable d'environnement GLOB_SALT_DOWN est vide : doit contenir une secret pour l'encryption. (remplir au pif)")
		response.IsConfigured = false
	}

	if len(env.AuthBlobAdmin) == 0 {
		response.Message = append(response.Message, "La variable d'environnement PW_HASH_ADMIN est vide : doit contenir le hash d'authentification l'admin (voir page de génération)")
		response.IsConfigured = false
	}

	if len(env.AuthBlobRegular) == 0 {
		response.Message = append(response.Message, "La variable d'environnement PW_HASH_USER est vide : doit contenir le hash d'authentification d'un utilisateur normal (voir page de génération)")
		response.IsConfigured = false
	}

	c.JSON(http.StatusOK, &response)
}

func (env *Env) RequestHash(c *gin.Context) {
	var req struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}

	err := c.BindJSON(&req)
	if err != nil {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.User+" / "+string(env.PasswordHashSalt)+" / "+req.Pass), 12)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error while hashing")
		return
	}

	var response struct {
		Hash string `json:"hash"`
	}

	response.Hash = string(hash)
	c.JSON(http.StatusOK, &response)
}

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
	var maxBandwidth int64 = 0
	var maxDuration int64 = 0
	err = bcrypt.CompareHashAndPassword(env.AuthBlobAdmin, []byte(req.User+" / "+string(env.PasswordHashSalt)+" / "+req.Pass))
	if err != nil {
		userAdmin = false
		maxBandwidth = env.NonAdminSpeedLimit
		maxDuration = env.NonAdminTimeLimit
		err = bcrypt.CompareHashAndPassword(env.AuthBlobRegular, []byte(req.User+" / "+string(env.PasswordHashSalt)+" / "+req.Pass))
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
		"max_duration":  maxDuration,
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
		"max_duration":  maxDuration,
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
