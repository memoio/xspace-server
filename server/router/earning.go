package router

import "github.com/gin-gonic/gin"

func loadEarningMoudles(r *gin.RouterGroup, h *handle) {
	r.GET("/info", h.earningInfo)
}

func (h *handle) earningInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
		"data": EarningInfo{
			Sum:       100,
			Activitys: []Activity{},
		},
	})
}
