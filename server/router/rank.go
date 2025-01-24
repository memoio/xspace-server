package router

import "github.com/gin-gonic/gin"

func loadRankMoudles(r *gin.RouterGroup, h *handle) {
	r.GET("/activitylist", h.RankActivityList)
	r.GET("/activityinfo", h.RankActivityInfo)
}

func (h *handle) RankActivityList(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
		"data":   []string{},
	})
}

func (h *handle) RankActivityInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
		"data":   []ActivityInfo{
			
		},
	})
}
