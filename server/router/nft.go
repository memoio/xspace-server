package router

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/types"
)

func LoadNFTModule(r *gin.RouterGroup, h *handler) {
	r.POST("/tweet/mint", h.VerifyIdentityHandler, h.mintTweet)
	r.POST("/data/mint", h.VerifyIdentityHandler, h.mintData)
	r.GET("/list", h.VerifyIdentityHandler, h.listNFT)
	r.GET("/tweet/info", h.VerifyIdentityHandler, h.twitterNFTInfo)
	r.GET("/data/info", h.VerifyIdentityHandler, h.dataNFTInfo)
}

// @ Summary MintTweet
//
//	@Description	Mint user's tweets into NFTs
//	@Tags			NFT
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			name			body		string	true	"User's twtter/x name"
//	@Param			postTime		body		int64	true	"The timestamp when the user posted the tweet"
//	@Param			tweet			body		string	true	"The text of the tweet(including emoji)"
//	@Param			image			body		string	true	"The image url of the tweet"
//	@Success		200				{object}	types.MintRes
//	@Router			/v1/nft/tweet/mint [post]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) mintTweet(c *gin.Context) {
	address := c.GetString("address")

	var req types.MintTweetReq
	err := c.BindJSON(&req)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithError(400, err)
		return
	}

	tokenId, err := h.nftController.MintTweetNFTTo(h.context, req.Name, req.PostTime, req.Tweet, req.Images, common.HexToAddress(address))
	if err != nil {
		h.logger.Error(err)
		c.AbortWithError(520, err)
	}
	c.JSON(200, types.MintRes{TokenID: tokenId})
}

// @ Summary MintData
//
//	@Description	Mint user's data into NFTs
//	@Tags			NFT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			file			formData	file	true	"User's data"
//	@Success		200				{object}	types.MintRes
//	@Router			/v1/nft/data/mint [post]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) mintData(c *gin.Context) {
	address := c.GetString("address")
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error(err)
		c.JSON(400, err)
		return
	}

	fr, err := file.Open()
	if err != nil {
		h.logger.Error(err)
		c.JSON(400, err)
		return
	}

	tokenId, err := h.nftController.MintDataNFTTo(h.context, file.Filename, fr, common.HexToAddress(address))
	if err != nil {
		h.logger.Error(err)
		c.JSON(520, err)
		return
	}

	c.JSON(200, types.MintRes{TokenID: tokenId})
}

// @ Summary ListNFT
//
//	@Description	List all NFT information belonging to the user
//	@Tags			NFT
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			page			query		string	true	"Pages"
//	@Param			size			query		string	true	"The amount of data displayed on each page"
//	@Param			type			query		string	false	"NFT type (tweet for tweetNFT, data for dataNFT, tweetNFT and dataNFT will be all listed by default)"
//	@Param			order			query		string	false	"Order rules (date_asc for sorting by creation time from smallest to largest, date_desc for sorting by creation time from largest to smallest)"
//	@Success		200				{object}	types.ListNFTRes
//	@Router			/v1/nft/list [get]
//	@Failure		400	{object}	error
//	@Failure		520	{object}	error
func (h *handler) listNFT(c *gin.Context) {
	address := c.GetString("address")
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	ntype := c.Query("type")
	order := c.Query("order")

	if order == "" {
		order = ""
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

	var nfts []database.NFTStore
	if ntype == "" {
		nfts, err = database.ListNFT(page, size, address, order)
	} else {
		nfts, err = database.ListNFTByType(page, size, address, order, ntype)
	}
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, types.ListNFTRes{nfts})
}

// @ Summary TwitterNFTInfo
//
//	@Description	Get TweetNFT content
//	@Tags			NFT
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			tokenID			query		string	true	"TweetNFT's id"
//	@Success		200				{object}	types.TweetNFTInfoRes
//	@Router			/v1/nft/tweet/info [get]
//	@Failure		400	{object}	error
//	@Failure		403	{object}	error
//	@Failure		520	{object}	error
func (h *handler) twitterNFTInfo(c *gin.Context) {
	address := c.GetString("address")
	tokenIdStr := c.Query("tokenID")

	tokenId, err := strconv.ParseUint(tokenIdStr, 10, 64)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	info, err := database.GetNFTInfo(tokenId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	if info.Address != address {
		h.logger.Error("You have no access to the tweet nft")
		c.AbortWithStatusJSON(403, "You have no access to the tweet nft")
		return
	}

	if info.Type != "tweet" {
		h.logger.Error("This api only support tweet NFT, but got type: " + info.Type)
		c.AbortWithStatusJSON(403, "This api only support tweet NFT, but got type: "+info.Type)
		return
	}

	content, err := h.nftController.GetTweetNFTContent(h.context, tokenId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	c.JSON(200, content)
}

// @ Summary DataNFTInfo
//
//	@Description	Get DataNFT content
//	@Tags			NFT
//	@Accept			json
//	@Produce		octet-stream
//	@Param			Authorization	header	string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			tokenID			query	string	true	"DataNFT's id"
//	@Success		200				{file}	binary	"DataNFT binary content"
//	@Router			/v1/nft/data/info [get]
//	@Failure		400	{object}	error
//	@Failure		403	{object}	error
//	@Failure		520	{object}	error
func (h *handler) dataNFTInfo(c *gin.Context) {
	address := c.GetString("address")
	tokenIdStr := c.Query("tokenID")

	tokenId, err := strconv.ParseUint(tokenIdStr, 10, 64)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(400, err)
		return
	}

	info, err := database.GetNFTInfo(tokenId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	if info.Address != address {
		h.logger.Error("You have no access to the data nft")
		c.AbortWithStatusJSON(403, "You have no access to the data nft")
		return
	}

	if info.Type != "data" {
		h.logger.Error("This api only support data NFT, but got type: " + info.Type)
		c.AbortWithStatusJSON(403, "This api only support data NFT, but got type: "+info.Type)
		return
	}

	contentInfo, data, err := h.nftController.GetDataNFTContent(h.context, tokenId)
	if err != nil {
		h.logger.Error(err)
		c.AbortWithStatusJSON(520, err)
		return
	}

	extraHeaders := map[string]string{
		"Content-Disposition": "attachment; filename=" + contentInfo.Name,
	}
	c.DataFromReader(200, contentInfo.Size, contentInfo.CType, data, extraHeaders)
}
