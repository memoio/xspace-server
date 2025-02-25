package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/types"
	"golang.org/x/xerrors"
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
//	@Failure		520	{object}	error
func (h *handler) userInfo(c *gin.Context) {
	address := c.GetString("address")

	user, err := database.GetUserInfo(address)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, types.UserInfoRes{Address: user.Address, InviteCode: user.InviteCode, Points: user.Points, Referrals: user.Referrals, Space: user.Space})
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
//	@Failure		520	{object}	error
func (h *handler) pointInfo(c *gin.Context) {
	address := c.GetString("address")

	user, err := database.GetUserInfo(address)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	godataCount, err := database.GetActionCount(address, 3)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	chargingCount, err := database.GetActionCount(address, 2)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, types.PointInfoRes{Points: user.Points, GodataCount: godataCount, GodataSpace: user.Space, ChargingCount: chargingCount})
}

// @ Summary Charge
//
//	@Description	Users can charge once every 6 hours
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Success		200				{object}	types.UserInfoRes
//	@Router			/v1/point/charge [post]
//	@Failure		403	{object}	error
func (h *handler) charge(c *gin.Context) {
	address := c.GetString("address")

	actions, err := database.ListActionHistoryByID(address, 1, 5, "date_desc", 2)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	if len(actions) >= 0 && actions[0].Time.Add(5*time.Hour).After(time.Now()) {
		err = xerrors.Errorf("The last charge time is %s, please try again after %s", actions[0].Time.String(), actions[0].Time.Add(5*time.Hour).String())
		h.logger.Error(err)
		c.AbortWithStatusJSON(403, err)
		return
	}

	user, err := h.pointController.FinishAction(address, 2)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, types.UserInfoRes{Address: user.Address, InviteCode: user.InviteCode, Points: user.Points, Referrals: user.Referrals, Space: user.Space})
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
//	@Param			order			query		string	false	"Order rules (date_asc for sorting by creation time from smallest to largest, date_desc for sorting by creation time from largest to smallest)"
//	@Success		200				{object}	types.PointHistoryRes
//	@Router			/v1/point/history [get]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) pointHistory(c *gin.Context) {
	address := c.GetString("address")
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	actionIdStr := c.Query("actionID")
	order := c.Query("order")

	if order == "" {
		order = "date_desc"
	}

	if actionIdStr == "" {
		actionIdStr = "-1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	actionId, err := strconv.Atoi(actionIdStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	actions, err := database.ListActionHistoryByID(address, page, size, order, actionId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, types.PointHistoryRes{History: append(actions, database.ActionStore{
		Id:      1,
		Name:    "Charging",
		Address: address,
		Point:   3,
		Time:    time.Now().Add(-4 * time.Hour),
	})})
}

// func (h *handler) invited(c *gin.Context) {

// }

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
