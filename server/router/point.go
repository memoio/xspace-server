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
	r.POST("/point/invite", h.VerifyIdentityHandler, h.invite)
	r.POST("/point/add", h.VerifyIdentityHandler, h.finishAction)
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
		c.AbortWithStatusJSON(520, err.Error())
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
		c.AbortWithStatusJSON(520, err.Error())
		return
	}

	godataCount, err := database.GetActionCount(address, 3)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err.Error())
		return
	}

	chargingCount, err := database.GetActionCount(address, 2)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err.Error())
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
//	@Failure		520	{object}	error
func (h *handler) charge(c *gin.Context) {
	address := c.GetString("address")

	user, err := h.pointController.FinishAction(address, 2)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithError(520, err)
		return
	}

	c.JSON(200, types.UserInfoRes{Address: user.Address, InviteCode: user.InviteCode, InvitedCode: user.InvitedCode, Points: user.Points, Referrals: user.Referrals, Space: user.Space})
}

// @ Summary FinishAction
//
//	@Description	Users can earn point with finish action
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			tokenId			body		int		true	"action id (101 for follow twitter, 102 for follow discord, 103 for follow telegram)"
//	@Success		200				{object}	types.UserInfoRes
//	@Router			/v1/point/add [post]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) finishAction(c *gin.Context) {
	address := c.GetString("address")

	var req types.FinishActionReq
	err := c.BindJSON(&req)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	if req.ActionId < 101 || req.ActionId > 103 {
		h.logger.Error("This interface only support actionId from 101 to 103")
		c.AbortWithStatusJSON(400, "This interface only support actionId from 101 to 103")
		return
	}

	user, err := h.pointController.FinishAction(address, req.ActionId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err.Error())
		return
	}

	c.JSON(200, types.UserInfoRes{Address: user.Address, InviteCode: user.InviteCode, InvitedCode: user.InvitedCode, Points: user.Points, Referrals: user.Referrals, Space: user.Space})
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
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	actionId, err := strconv.Atoi(actionIdStr)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	actions, err := database.ListActionHistoryByID(address, page, size, order, actionId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err.Error())
		return
	}

	c.JSON(200, types.PointHistoryRes{History: actions})
}

// @ Summary Invite
//
//	@Description	Get the history of the point info by address
//	@Tags			Point
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			code			body		string	true	"The invite code from other user"
//	@Success		200				{object}	types.UserInfoRes
//	@Router			/v1/point/invite [post]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) invite(c *gin.Context) {
	address := c.GetString("address")

	var req types.InviteReq
	err := c.BindJSON(&req)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	if len(req.Code) != 6 {
		h.logger.Error("The invitation code must be 6 characters long")
		c.AbortWithStatusJSON(400, "The invitation code must be 6 characters long")
		return
	}

	user1, err := database.GetUserInfo(address)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	if user1.InviteCode == req.Code {
		h.logger.Error("You can't invite yourself")
		c.AbortWithStatusJSON(400, "You can't invite yourself")
		return
	}

	user2, err := database.GetUserInfoByCode(req.Code)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	_, err = h.pointController.FinishAction(user2.Address, 12)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	user, err := h.pointController.FinishAction(user1.Address, 11)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err.Error())
		return
	}

	user.InvitedCode = req.Code
	err = user.UpdateUserInfo()
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err.Error())
		return
	}

	c.JSON(200, types.UserInfoRes{Address: user.Address, InviteCode: user.InviteCode, InvitedCode: user.InvitedCode, Points: user.Points, Referrals: user.Referrals, Space: user.Space})
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
	c.JSON(200, types.ListProjectsRes{Projects: []types.ProjectInfo{{Name: "Data-Did", ProjectID: 1, Start: time.Now(), End: time.Now().Add(96 * time.Hour)}}})
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
	c.JSON(200, types.RankRes{RankInfo: []types.RankInfo{{Rank: 1, Address: "0xcFA4816BE86B7b56A5373A36bE5B9c53c0f157f8", Scores: 100000, Points: 10000}}})
}
