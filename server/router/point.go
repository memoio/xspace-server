package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/types"
)

func LoadPointModules(r *gin.RouterGroup, h *handler) {
	r.GET("/user/info", h.VerifyIdentityHandler, h.userInfo)

	r.POST("/point/charge", h.VerifyIdentityHandler, h.charge)
	r.GET("/point/info", h.VerifyIdentityHandler, h.pointInfo)
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
//	@Success		200				{object}	types.UserInfoRes
//	@Router			/v1/user/info [get]
//	@Failure		500	{object}	error
func (h *handler) userInfo(c *gin.Context) {
	c.JSON(200, types.UserInfoRes{Address: "0x", InviteCode: "78ED", Points: 300, Referrals: 2, Space: 5})
}

// @ Summary UserInfo
//
//	@Description	Get the user's point info by address
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{object}	types.PointInfoRes
//	@Router			/v1/point/info [get]
//	@Failure		500	{object}	error
func (h *handler) pointInfo(c *gin.Context) {
	c.JSON(200, types.PointInfoRes{Points: 3000, GodataCount: 5, GodataSpace: 97, ChargingCount: 2, Charging: true})
}

// @ Summary Charge
//
//	@Description	Users can charge once every 6 hours
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{object}	types.PointInfoRes
//	@Router			/v1/point/charge [post]
//	@Failure		500	{object}	error
func (h *handler) charge(c *gin.Context) {
	c.JSON(200, types.PointInfoRes{Points: 3000, GodataCount: 5, ChargingCount: 2})
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
//	@Param			actionID		query		string	false	"The action id"
//	@Param			order			query		string	false	"Order rules (date_asc for sorting by creation time from smallest to largest, date_dsc for sorting by creation time from largest to smallest)"
//	@Success		200				{object}	types.PointHistoryRes
//	@Router			/v1/point/history [get]
//	@Failure		500	{object}	error
//	@Failure		501	{object}	error
func (h *handler) pointHistory(c *gin.Context) {
	address := c.GetString("address")
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	actionIdStr := c.Query("actionID")
	order := c.Query("order")

	if order == "" {
		order = "date_dsc"
	}

	if actionIdStr == "" {
		actionIdStr = "-1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(500, err)
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(500, err)
		return
	}

	actionId, err := strconv.Atoi(actionIdStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(500, err)
		return
	}

	actions, err := database.ListActionHistoryByID(address, page, size, order, actionId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(501, err)
		return
	}

	c.JSON(200, types.PointHistoryRes{History: append(actions, database.ActionStore{
		Id:      1,
		Name:    "Charging",
		Address: address,
		Time:    time.Now().Add(-4 * time.Hour),
	})})
}

// @ Summary ListProjects
//
//	@Description	List all projects with Xspace
//	@Tags			Rank
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	types.ListProjectsRes
//	@Router			/v1/project/list [get]
//	@Failure		500	{object}	error
func (h *handler) listProjects(c *gin.Context) {
	c.JSON(200, types.ListProjectsRes{[]types.ProjectInfo{types.ProjectInfo{Name: "Data-Did", ProjectID: 1, Start: time.Now(), End: time.Now().Add(96 * time.Hour)}}})
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
//	@Success		200		{object}	types.RankRes
//	@Router			/v1/project/rank [get]
//	@Failure		500	{object}	error
func (h *handler) rank(c *gin.Context) {
	c.JSON(200, types.RankRes{[]types.RankInfo{types.RankInfo{Rank: 1, Address: "0xcFA4816BE86B7b56A5373A36bE5B9c53c0f157f8", Scores: 100000, Points: 10000}}})
}
