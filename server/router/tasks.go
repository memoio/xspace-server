package router

import "github.com/gin-gonic/gin"

func loadTasksMoudles(r *gin.RouterGroup, h *handle) {
	r.POST("/add", h.tasksAdd)
}

func (h *handle) tasksAdd(c *gin.Context) {
	c.JSON(200, gin.H{
		"result": 1,
	})
}
