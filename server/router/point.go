package router

import (
	"time"

	"github.com/gin-gonic/gin"
)

func LoadPointModules(r *gin.RouterGroup, h *handler) {
	r.GET("/user/info", h.VerifyIdentityHandler, h.pointInfo)

	r.POST("/point/charge", h.VerifyIdentityHandler, h.charge)
	r.GET("/point/history", h.VerifyIdentityHandler, h.pointHistory)

	r.GET("/project/list", h.listProjects)
	r.GET("/project/rank", h.rank)
}

// @ Summary UserInfo
//
//	@Description	Get the user basic info by address
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{object}	PointInfoRes
//	@Router			/v1/user/info [get]
//	@Failure		500	{object}	error
func (h *handler) pointInfo(c *gin.Context) {
	c.JSON(200, PointInfoRes{Points: 3000, GodataCount: 5, GodataSpace: 97, ChargingCount: 2, Charging: true})
}

// @ Summary Charge
//
//	@Description	Users can charge once every 6 hours
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{object}	PointInfoRes
//	@Router			/v1/point/charge [post]
//	@Failure		500	{object}	error
func (h *handler) charge(c *gin.Context) {
	c.JSON(200, PointInfoRes{Points: 3000, GodataCount: 5, ChargingCount: 2})
}

// @ Summary PointHistory
//
//	@Description	Get the history of the point info by address
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			page			query		string	true	"Pages"
//	@Param			size			query		string	true	"The amount of data displayed on each page"
//	@Param			order			query		string	false	"Order rules (date_asc for sorting by creation time from smallest to largest, date_dsc for sorting by creation time from largest to smallest)"
//	@Success		200				{object}	PointHistoryRes
//	@Router			/v1/point/history [get]
//	@Failure		500	{object}	error
func (h *handler) pointHistory(c *gin.Context) {
	c.JSON(200, PointHistoryRes{[]PointInfo{PointInfo{Point: 50, Time: time.Now(), ActionName: "daily sign-in"}}})
}

// @ Summary ListProjects
//
//	@Description	List all projects with Xspace
//	@Tags			Rank
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ListProjectsRes
//	@Router			/v1/project/list [get]
//	@Failure		500	{object}	error
func (h *handler) listProjects(c *gin.Context) {
	c.JSON(200, ListProjectsRes{[]ProjectInfo{ProjectInfo{Name: "Data-Did", ProjectID: 1, Start: time.Now(), End: time.Now().Add(96 * time.Hour)}}})
}

// @ Summary Rank
//
//	@Description	Get the ranking of cooperative projects
//	@Tags			Rank
//	@Accept			json
//	@Produce		json
//	@Param			id		query		string	true	"cooperative project id"
//	@Param			page	query		string	true	"Pages"
//	@Param			size	query		string	true	"The amount of data displayed on each page"
//	@Success		200		{object}	RankRes
//	@Router			/v1/project/rank [get]
//	@Failure		500	{object}	error
func (h *handler) rank(c *gin.Context) {
	c.JSON(200, RankRes{[]RankInfo{RankInfo{Rank: 1, Address: "0xcFA4816BE86B7b56A5373A36bE5B9c53c0f157f8", Scores: 100000, Points: 10000}}})
}
