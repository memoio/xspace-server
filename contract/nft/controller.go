package nft

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	com "github.com/memoio/contractsv2/common"
	"github.com/memoio/nft-solidity/go-contracts/token"
	"github.com/memoio/xspace-server/database"
	"github.com/memoio/xspace-server/gateway"
	"github.com/memoio/xspace-server/point"
	"github.com/memoio/xspace-server/types"
	"golang.org/x/xerrors"
)

var (
	tweetBucket = "tweet-nft"
	dataBucket  = "data-nft"

	checkTxSleepTime = 6 // 先等待6s（出块时间加1）
	nextBlockTime    = 5 // 出块时间5s
)

type NFTType string

const (
	TweetNFT NFTType = "tweet"
	DataNFT  NFTType = "data"
)

type NFTController struct {
	contractAddress common.Address
	endpoint        string
	transactor      *bind.TransactOpts
	store           gateway.IGateway
	pointController *point.PointController
	logger          *log.Helper
}

func NewNFTController(contractAddress common.Address, endpoint, sk string, logger *log.Helper) (*NFTController, error) {
	store, err := gateway.NewGateway(logger)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(sk)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	client, err := ethclient.DialContext(context.TODO(), endpoint)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	auth.Value = big.NewInt(0) // in wei

	pointController, err := point.NewPointController()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &NFTController{
		contractAddress: contractAddress,
		endpoint:        endpoint,
		transactor:      auth,
		store:           store,
		pointController: pointController,
		logger:          logger,
	}, nil
}

func (c *NFTController) MintDataNFT(ctx context.Context, filename string, data io.Reader) (uint64, error) {
	return c.MintDataNFTTo(ctx, filename, data, c.transactor.From)
}

func (c *NFTController) MintDataNFTTo(ctx context.Context, filename string, data io.Reader, to common.Address) (uint64, error) {
	return c.mintNFTTo(ctx, DataNFT, filename, data, to)
}

func (c *NFTController) MintTweetNFT(ctx context.Context, name string, postTime int64, tweet string, images []string, link string) (uint64, error) {
	return c.MintTweetNFTTo(ctx, name, postTime, tweet, images, link, c.transactor.From)
}

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

func (c *NFTController) mintNFTTo(ctx context.Context, ntype NFTType, filename string, r io.Reader, to common.Address) (uint64, error) {
	var bucket string
	if ntype == TweetNFT {
		bucket = tweetBucket
	} else if ntype == DataNFT {
		bucket = dataBucket
	} else {
		return 0, xerrors.Errorf("unspported nft type: %s", ntype)
	}

	userInfo, err := database.GetUserInfo(to.Hex())
	if err != nil {
		return 0, err
	}

	if userInfo.Space == 0 {
		return 0, xerrors.New("The user's current storage units is 0")
	}

	info, err := c.store.PutObject(ctx, bucket, filename, r, gateway.ObjectOptions{})
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	client, err := ethclient.DialContext(ctx, c.endpoint)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	nftIns, err := token.NewERC721(c.contractAddress, client)
	if err != nil {
		return 0, err
	}

	tokenId, err := nftIns.Id(&bind.CallOpts{})
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	c.logger.Info(tokenId)
	tx, err := nftIns.Mint(c.transactor, to, string(ntype)+`\`+info.Cid)
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	err = c.checkTx(tx.Hash(), "mint")
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	nftStore := &database.NFTStore{
		TokenId: tokenId.Uint64(),
		Address: to.Hex(),
		Cid:     info.Cid,
		Type:    string(ntype),
		Time:    time.Now(),
	}
	err = nftStore.CreateNFTInfo()
	if err != nil {
		return tokenId.Uint64(), err
	}

	_, err = c.pointController.FinishAction(to.Hex(), 3)
	return tokenId.Uint64(), err
}

func (c *NFTController) GetDataNFTContent(ctx context.Context, tokenId uint64) (gateway.ObjectInfo, io.Reader, error) {
	nftType, info, r, err := c.getNFTContent(ctx, tokenId)
	if err != nil {
		return gateway.ObjectInfo{}, nil, err
	}

	if nftType != TweetNFT {
		return gateway.ObjectInfo{}, nil, xerrors.Errorf(`got wrong nft type: %s`, nftType)
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

func (c *NFTController) getNFTContent(ctx context.Context, tokenId uint64) (NFTType, gateway.ObjectInfo, io.Reader, error) {
	var nftType NFTType
	client, err := ethclient.DialContext(ctx, c.endpoint)
	if err != nil {
		return nftType, gateway.ObjectInfo{}, nil, err
	}
	defer client.Close()

	nftIns, err := token.NewERC721(c.contractAddress, client)
	if err != nil {
		return nftType, gateway.ObjectInfo{}, nil, err
	}

	tokenUri, err := nftIns.TokenURI(&bind.CallOpts{}, big.NewInt(int64(tokenId)))
	if err != nil {
		return nftType, gateway.ObjectInfo{}, nil, err
	}

	splits := strings.Split(tokenUri, `\`)
	if len(splits) != 2 {
		return nftType, gateway.ObjectInfo{}, nil, xerrors.Errorf("can't resolve token uri: %d", tokenUri)
	}
	nftType = NFTType(splits[0])
	cid := splits[1]

	var buffer bytes.Buffer
	err = c.store.GetObject(ctx, cid, &buffer, gateway.ObjectOptions{})
	if err != nil {
		return nftType, gateway.ObjectInfo{}, nil, err
	}

	objInfo, err := c.store.GetObjectInfo(ctx, cid)

	return nftType, objInfo, &buffer, err
}

func (c *NFTController) checkTx(txHash common.Hash, name string) error {
	var receipt *etypes.Receipt

	t := checkTxSleepTime
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(t) * time.Second)
		receipt = com.GetTransactionReceipt(c.endpoint, txHash)
		if receipt != nil {
			break
		}
		t = nextBlockTime
	}

	if receipt == nil {
		err := xerrors.Errorf("%s: cann't get transaction(%s) receipt, not packaged", name, txHash)
		c.logger.Error(err)
		return err
	}

	// 0 means fail
	if receipt.Status == 0 {
		if receipt.GasUsed != receipt.CumulativeGasUsed {
			err := xerrors.Errorf("%s: transaction(%s) exceed gas limit", name, txHash)
			c.logger.Error(err)
			return err
		}

		err := xerrors.Errorf("%s: transaction(%s) mined but execution failed, please check your tx input", name, txHash)
		c.logger.Error(err)
		return err
	}
	return nil
}

// func getStorageUnits(address common.Address) (int, error) {
// 	userInfo, err := database.GetUserInfo(address.Hex())
// 	if err != nil {
// 		return 0, err
// 	}

// 	if userInfo.UpdateTime.Add(24 * time.Hour).Before(time.Now()) {
// 		userInfo.Space = config.DefaultSpace
// 		userInfo.UpdateTime = time.Now()
// 		err = userInfo.UpdateUserInfo()
// 		if err != nil {
// 			return 0, err
// 		}
// 	}

// 	return userInfo.Space, nil
// }

// func finishMint(c *point.PointController, address common.Address) error {
// 	userInfo, err := database.GetUserInfo(address.Hex())
// 	if err != nil {
// 		return err
// 	}

// 	if userInfo.Space == 0 {
// 		return xerrors.New("The user's current storage units is 0")
// 	}

// 	actionInfo, err := c.GetActionInfo(3)
// 	if err != nil {
// 		return err
// 	}

// 	userInfo.Space -= 1
// 	userInfo.Points += actionInfo.Point
// 	err = userInfo.UpdateUserInfo()
// 	if err != nil {
// 		return err
// 	}

// 	action := database.ActionStore{
// 		Id:      actionInfo.ID,
// 		Name:    actionInfo.Name,
// 		Address: address.Hex(),
// 		Point:   actionInfo.Point,
// 		Time:    time.Now(),
// 	}

// 	return action.CreateActionInfo()
// }
