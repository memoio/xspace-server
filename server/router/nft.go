package router

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
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
//	@Param			postTime		body		string	true	"The time when the user posted the tweet"
//	@Param			tweet			body		string	true	"The text of the tweet(including emoji)"
//	@Param			image			body		string	true	"The image url of the tweet"
//	@Success		200				{object}	MintRes
//	@Router			/v1/nft/mint [post]
//	@Failure		502	{object}	error
//	@Failure		503	{object}	error
func (h *handler) mintTweet(c *gin.Context) {
	c.JSON(200, MintRes{TokenID: 100})
}

// @ Summary MintData
//
//	@Description	Mint user's data into NFTs
//	@Tags			NFT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			file			formData	file	true	"User's data"
//	@Success		200				{object}	MintRes
//	@Router			/v1/nft/mint [post]
//	@Failure		502	{object}	error
//	@Failure		503	{object}	error
func (h *handler) mintData(c *gin.Context) {
	c.JSON(200, MintRes{TokenID: 100})
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
//	@Param			type			query		string	false	"NFT type (1 for tweetNFT, 2 for dataNFT, tweetNFT and dataNFT will be all listed by default)"
//	@Param			order			query		string	false	"Order rules (date_asc for sorting by creation time from smallest to largest, date_dsc for sorting by creation time from largest to smallest)"
//	@Success		200				{object}	ListNFTRes
//	@Router			/v1/nft/list [get]
//	@Failure		500	{object}	error
func (h *handler) listNFT(c *gin.Context) {
	c.JSON(200, ListNFTRes{[]NFTInfo{NFTInfo{TokenID: 100, Type: 1, CreateTime: time.Now()}}})
}

// @ Summary TwitterNFTInfo
//
//	@Description	Get TweetNFT content
//	@Tags			NFT
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer YOUR_ACCESS_TOKEN"
//	@Param			tokenID			query		string	true	"TweetNFT's id"
//	@Success		200				{object}	TweetNFTInfoRes
//	@Router			/v1/nft/tweet/info [get]
//	@Failure		500	{object}	error
func (h *handler) twitterNFTInfo(c *gin.Context) {
	c.JSON(200, TweetNFTInfoRes{Name: "test", PostTime: time.Now().Unix(), Tweet: "hello, twitter.😀", Images: []string{}})
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
//	@Failure		500	{object}	error
func (h *handler) dataNFTInfo(c *gin.Context) {
	var w bytes.Buffer
	w.WriteString("hello,world\n")
	extraHeaders := map[string]string{
		"Content-Disposition": "attachment; filename=\"test\"",
	}
	c.DataFromReader(200, int64(w.Len()), "txt", &w, extraHeaders)
}
