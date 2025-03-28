package nft

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/nft/contract"
	"github.com/memoio/xspace-server/nft/storage"
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
	store       storage.IGateway
	nftContract *contract.NFTContract
	logger      *log.Helper
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

	return &NFTController{
		store:       store,
		nftContract: contract,
		logger:      logger,
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
