package router

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	auth "github.com/memoio/xspace-server/authentication"
	"golang.org/x/xerrors"
)

func LoadAuthModule(g *gin.RouterGroup, h *handler) {
	g.GET("/challenge", h.ChallengeHandler())

	g.POST("/login", h.LoginHandler())

	g.GET("/refresh", h.RefreshHandler())

	g.GET("/identity", h.VerifyIdentityHandler, func(c *gin.Context) {
		c.JSON(200, gin.H{
			"address": c.GetString("address"),
			"chainid": c.GetInt("chainid"),
		})
	})
}

// @ Summary Challenge
//
//	@Description	Get the challenge message by address before you login
//	@Tags			Login
//	@Accept			json
//	@Produce		json
//	@Param			address	query		string	true	"User's address (connect to xspace)"
//	@Param			chainid	query		string	true	"The network ID which the user's wallet is connected to"
//	@Param			Origin	header		string	true	"The frontend's domain"
//	@Success		200		{string}	string	"The challenge message"
//	@Router			/v1/challenge [get]
//	@Failure		400	{object}	error
func (h *handler) ChallengeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Query("address")
		uri, err := url.Parse(c.GetHeader("Origin"))
		if err != nil {
			h.logger.Error(err)
			c.AbortWithStatusJSON(400, err.Error())
			return
		}
		domain := uri.Host

		var chainID int
		if c.Query("chainid") != "" {
			chainID, err = strconv.Atoi(c.Query("chainid"))
			if err != nil {
				h.logger.Error(err)
				c.AbortWithStatusJSON(400, err.Error())
				return
			}
		} else {
			chainID = 985
		}

		challenge, err := h.authController.Challenge(domain, address, uri.String(), chainID)
		if err != nil {
			h.logger.Error(err)
			c.AbortWithStatusJSON(400, err.Error())
			return
		}
		c.String(http.StatusOK, challenge)
	}
}

// @ Summary Login
//
//	@Description	Use the signMessage method to sign the challenge message. After signing, call the login interface to complete the login.
//	@Description	If the login is successful, the Login API will return an Access Token and a Refresh Token. When accessing subsequent APIs, you need to add the Authorization field in the headers with the value "Bearer Your_Access_Token"
//	@Tags			Login
//	@Accept			json
//	@Produce		json
//	@Param			message		body		string				true	"The challenge message"
//	@Param			signature	body		string				true	"The result after the user's private key signs the challenge message"
//	@Success		200			{object}	map[string]string	"The access token and refresh token"
//	@Router			/v1/login [post]
//	@Failure		400	{object}	error
//	@Failure		401	{object}	error
func (h *handler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request auth.EIP4361Request
		err := c.BindJSON(&request)
		if err != nil {
			h.logger.Error(err)
			c.AbortWithStatusJSON(400, err.Error())
			return
		}
		accessToken, refreshToken, err := h.authController.Login(request)
		if err != nil {
			h.logger.Error(err)
			c.AbortWithStatusJSON(401, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		})
	}
}

// @ Summary Refresh
//
//	@Description	If the access token expires, you can call the refresh API to get a new access token or log in again.
//	@Tags			Login
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer YOUR_FRESH_TOKEN"
//	@Success		200				{object}	map[string]string	"The access token"
//	@Router			/v1/refresh [post]
//	@Failure		401	{object}	error
func (h *handler) RefreshHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		accessToken, err := auth.VerifyRefreshToken(tokenString)
		if err != nil {
			h.logger.Error(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, xerrors.Errorf("Illegal refresh token: %s", err.Error()).Error())
			return
		}

		c.JSON(http.StatusOK, map[string]string{
			"accessToken": accessToken,
		})
	}
}

func (h *handler) VerifyIdentityHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		tokenString = "Bearer " + c.Query("token")
	}

	address, chainid, err := auth.VerifyAccessToken(tokenString)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(401, err.Error())
		return
	}

	c.Set("address", address)
	c.Set("chainid", chainid)
}
