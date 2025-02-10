package router

import "github.com/gin-gonic/gin"

func LoadReferModule(r *gin.RouterGroup, h *handler) {
	r.GET("/code", h.VerifyIdentityHandler, h.getReferCode)
	r.POST("/bind", h.VerifyIdentityHandler, h.bindReferCode)
}

// @ Summary ReferCode
//
//	@Description	Get the user's refer code
//	@Tags			Refer
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{string}	string	"user's refer code"
//	@Router			/v1/refer/info [get]
//	@Failure		500	{object}	error
func (h *handler) getReferCode(c *gin.Context) {
	c.JSON(200, "6YD8F9")
}

// @ Summary BindReferCode
//
//	@Description	Bind the refer code when first log in
//	@Tags			Refer
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			code			body		string	true	"Other user's refer code"
//	@Success		200				{string}	string
//	@Router			/v1/refer/bind [post]
//	@Failure		500	{object}	error
func (h *handler) bindReferCode(c *gin.Context) {
	c.JSON(200, "success")
}
