package router

import "github.com/gin-gonic/gin"

func loadReferMoudles(r *gin.RouterGroup, h *handle) {
	r.GET("/info", h.referInfo)
}

func (h *handle) referInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
		"data": UserInfo{
			Invitees: 5,
			Code:     "EfsDw",
			Reward:   200,
		},
	})
}
