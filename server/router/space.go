package router

import "github.com/gin-gonic/gin"

func loadSpaceMoudles(r *gin.RouterGroup, h *handle) {
	r.GET("/list", h.getSpaceList)
	r.GET("/info", h.getSpaceInfo)
}

func (h *handle) getSpaceList(c *gin.Context) {
	address := c.Query("address")
	types := c.Query("type")
	date := c.Query("date")

	h.logger.Info("getSpaceList", address, types, date)
	c.JSON(200, gin.H{
		"result": 1,
		"data":   []SpaceListInfo{},
	})
}

func (h *handle) getSpaceInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
		"data":   []byte{},
	})
}
