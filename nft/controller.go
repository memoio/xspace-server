package nft

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/nft/contract"
	"github.com/memoio/xspace-server/nft/storage"
	"github.com/memoio/xspace-server/point"
	"github.com/memoio/xspace-server/types"
	"golang.org/x/xerrors"
)

const (
	tweetBucket string        = "tweet-nft"
	dataBucket  string        = "data-nft"
	TweetNFT    types.NFTType = "tweet"
	DataNFT     types.NFTType = "data"
)

type NFTController struct {
	store           storage.IGateway
	nftContract     *contract.NFTContract
	pointController *point.PointController
	logger          *log.Helper
}

func NewNFTController(contractAddress common.Address, endpoint, sk string, logger *log.Helper) (*NFTController, error) {
	store, err := storage.NewGateway(logger)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	contract, err := contract.NewNFTContract(contractAddress, endpoint, sk, logger)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	pointController, err := point.NewPointController()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &NFTController{
		store:           store,
		nftContract:     contract,
		pointController: pointController,
		logger:          logger,
	}, nil
}

func (c *NFTController) Start(ctx context.Context) {
	c.nftContract.Start(ctx)
}

func (c *NFTController) Stop() error {
	return c.nftContract.Stop()
}

// func (c *NFTController) MintDataNFT(ctx context.Context, filename string, data io.Reader) (uint64, error) {
// 	return c.MintDataNFTTo(ctx, filename, data, c.transactor.From)
// }

func (c *NFTController) MintDataNFTTo(ctx context.Context, filename string, data io.Reader, to common.Address) (uint64, error) {
	return c.mintNFTTo(ctx, DataNFT, filename, data, to)
}

// func (c *NFTController) MintTweetNFT(ctx context.Context, name string, postTime int64, tweet string, images []string, link string) (uint64, error) {
// 	return c.MintTweetNFTTo(ctx, name, postTime, tweet, images, link, c.transactor.From)
// }

func (c *NFTController) MintTweetNFTTo(ctx context.Context, name string, postTime int64, tweet string, images []string, link string, to common.Address) (uint64, error) {
	data, err := json.Marshal(map[string]any{
		"name":     name,
		"postTime": postTime,
		"tweet":    tweet,
		"images":   images,
		"link":     link,
	})
	if err != nil {
		return 0, err
	}

	var dataBuffer = bytes.NewBuffer(data)
	filename := name + hex.EncodeToString(crypto.Keccak256(data))
	return c.mintNFTTo(ctx, TweetNFT, filename, dataBuffer, to)
}

func (c *NFTController) StoreTweetTo(ctx context.Context, name string, postTime int64, tweet string, images []string, link string, to common.Address) (string, error) {
	data, err := json.Marshal(map[string]any{
		"name":     name,
		"postTime": postTime,
		"tweet":    tweet,
		"images":   images,
		"link":     link,
	})
	if err != nil {
		return "", err
	}

	var dataBuffer = bytes.NewBuffer(data)
	filename := name + hex.EncodeToString(crypto.Keccak256(data))

	userInfo, err := database.GetUserInfo(to.Hex())
	if err != nil {
		return "", err
	}

	if userInfo.Storage == 0 {
		return "", xerrors.New("The user's current storage units is 0")
	}

	cid, err := c.storeData(ctx, TweetNFT, filename, dataBuffer, to)
	if err != nil {
		return "", err
	}

	nftStore := &database.NFTStore{
		TokenId: 0,
		Address: to.Hex(),
		Cid:     cid,
		Type:    string(TweetNFT),
		Time:    time.Now(),
	}
	err = nftStore.CreateNFTInfo()
	if err != nil {
		return cid, err
	}

	_, err = c.pointController.FinishAction(to.Hex(), 4)
	return cid, err
}

func (c *NFTController) mintNFTTo(ctx context.Context, ntype types.NFTType, filename string, r io.Reader, to common.Address) (uint64, error) {
	userInfo, err := database.GetUserInfo(to.Hex())
	if err != nil {
		return 0, err
	}

	if userInfo.Space == 0 || userInfo.Storage == 0 {
		return 0, xerrors.New("The user's current storage units is 0")
	}

	cid, err := c.storeData(ctx, TweetNFT, filename, r, to)
	if err != nil {
		return 0, err
	}

	return c.nftContract.AddMintNFTTask(ntype, cid, to)
}

func (c *NFTController) storeData(ctx context.Context, ntype types.NFTType, name string, r io.Reader, to common.Address) (string, error) {
	var bucket string
	if ntype == TweetNFT {
		bucket = tweetBucket
	} else if ntype == DataNFT {
		bucket = dataBucket
	} else {
		return "", xerrors.Errorf("unspported nft type: %s", ntype)
	}

	userInfo, err := database.GetUserInfo(to.Hex())
	if err != nil {
		return "", err
	}

	if userInfo.Space == 0 {
		return "", xerrors.New("The user's current storage units is 0")
	}

	info, err := c.store.PutObject(ctx, bucket, name, r, storage.ObjectOptions{})
	if err != nil {
		c.logger.Error(err)
		return "", err
	}

	return info.Cid, nil
}

func (c *NFTController) GetDataNFTContent(ctx context.Context, tokenId uint64) (storage.ObjectInfo, io.Reader, error) {
	nftType, info, r, err := c.getNFTContent(ctx, tokenId)
	if err != nil {
		return storage.ObjectInfo{}, nil, err
	}

	if nftType != TweetNFT {
		return storage.ObjectInfo{}, nil, xerrors.Errorf(`got wrong nft type: %s`, nftType)
	}

	return info, r, nil
}

func (c *NFTController) GetTweetNFTContent(ctx context.Context, tokenId uint64) (types.TweetNFTInfo, error) {
	var res types.TweetNFTInfo

	nftType, _, r, err := c.getNFTContent(ctx, tokenId)
	if err != nil {
		return res, err
	}

	if nftType != TweetNFT {
		return res, xerrors.Errorf(`got wrong nft type: %s`, nftType)
	}

	var buffer = new(bytes.Buffer)
	_, err = buffer.ReadFrom(r)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(buffer.Bytes(), &res)
	if err != nil {
		return res, err
	}

	if res.Link == "" {
		res.Link = "https://x.com/" + res.Name
	}
	return res, nil
}

func (c *NFTController) GetTweetContent(ctx context.Context, cid string) (types.TweetNFTInfo, error) {
	var res types.TweetNFTInfo
	var buffer bytes.Buffer
	err := c.store.GetObject(ctx, cid, &buffer, storage.ObjectOptions{})
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(buffer.Bytes(), &res)
	if err != nil {
		return res, err
	}

	if res.Link == "" {
		res.Link = "https://x.com/" + res.Name
	}
	return res, nil
}

func (c *NFTController) getNFTContent(ctx context.Context, tokenId uint64) (types.NFTType, storage.ObjectInfo, io.Reader, error) {
	var nftType types.NFTType
	tokenUri, err := c.nftContract.TokenURI(ctx, tokenId)
	if err != nil {
		return nftType, storage.ObjectInfo{}, nil, err
	}

	splits := strings.Split(tokenUri, `\`)
	if len(splits) != 2 {
		return nftType, storage.ObjectInfo{}, nil, xerrors.Errorf("can't resolve token uri: %d", tokenUri)
	}
	nftType = types.NFTType(splits[0])
	cid := splits[1]

	var buffer bytes.Buffer
	err = c.store.GetObject(ctx, cid, &buffer, storage.ObjectOptions{})
	if err != nil {
		return nftType, storage.ObjectInfo{}, nil, err
	}

	objInfo, err := c.store.GetObjectInfo(ctx, cid)

	return nftType, objInfo, &buffer, err
}
