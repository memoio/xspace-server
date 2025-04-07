package contract

import (
	"context"
	"math/big"
	"sync"
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
	"github.com/memoio/xspace-server/point"
	"github.com/memoio/xspace-server/types"
	"golang.org/x/xerrors"
)

var (
	checkTxSleepTime = 6 // 先等待6s（出块时间加1）
	nextBlockTime    = 5 // 出块时间5s
)

type NFTTask struct {
	Type    types.NFTType
	Cid     string
	To      common.Address
	TokenId uint64
}

type NFTContract struct {
	contractAddress common.Address
	endpoint        string
	transactor      *bind.TransactOpts
	nftTasks        []NFTTask
	tokenID         uint64
	modifyMutex     sync.Mutex
	pointController *point.PointController
	done            bool
	close           context.CancelFunc
	logger          *log.Helper
}

func NewNFTContract(contractAddress common.Address, endpoint, sk string, logger *log.Helper) (*NFTContract, error) {
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

	nftIns, err := token.NewERC721(contractAddress, client)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	tokenId, err := nftIns.Id(&bind.CallOpts{})
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &NFTContract{
		contractAddress: contractAddress,
		endpoint:        endpoint,
		transactor:      auth,
		tokenID:         tokenId.Uint64(),
		pointController: pointController,

		done:   false,
		close:  nil,
		logger: logger,
	}, nil
}

func (c *NFTContract) Start(ctx context.Context) {
	cctx, cancel := context.WithCancel(ctx)
	c.close = cancel

	go c.runNFTTask(cctx)
}

func (c *NFTContract) Stop() error {
	if c.close == nil {
		return xerrors.New("Not started yet")
	}

	c.close()
	for !c.done {
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (c *NFTContract) AddMintNFTTask(ntype types.NFTType, cid string, to common.Address) (uint64, error) {
	c.modifyMutex.Lock()
	defer c.modifyMutex.Unlock()

	c.tokenID = c.tokenID + 1
	c.nftTasks = append(c.nftTasks, NFTTask{
		Type:    ntype,
		Cid:     cid,
		To:      to,
		TokenId: c.tokenID,
	})

	return c.tokenID, nil
}

func (c *NFTContract) TokenURI(ctx context.Context, tokenId uint64) (string, error) {
	client, err := ethclient.DialContext(ctx, c.endpoint)
	if err != nil {
		return "", err
	}
	defer client.Close()

	nftIns, err := token.NewERC721(c.contractAddress, client)
	if err != nil {
		return "", err
	}

	return nftIns.TokenURI(&bind.CallOpts{}, big.NewInt(int64(tokenId)))
}

func (c *NFTContract) runNFTTask(ctx context.Context) {
	for {
		if len(c.nftTasks) > 0 {
			c.modifyMutex.Lock()
			task := c.nftTasks[0]
			c.nftTasks = c.nftTasks[1:]
			c.modifyMutex.Unlock()

			var tokenId uint64
			for err := xerrors.New("new error"); err != nil; {
				tokenId, err = c.mintNFTTo(ctx, task.Type, task.Cid, task.To)
				if err != nil {
					c.logger.Error(err)
				}
			}

			if tokenId > task.TokenId {
				c.modifyMutex.Lock()
				c.tokenID = c.tokenID + tokenId - task.TokenId
				c.modifyMutex.Unlock()
			}
			c.logger.Infof("mint success, tokenID: %d, cid: %s", task.TokenId, task.Cid)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}
}

func (c *NFTContract) mintNFTTo(ctx context.Context, ntype types.NFTType, cid string, to common.Address) (uint64, error) {
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
	tx, err := nftIns.Mint(c.transactor, to, string(ntype)+`\`+cid)
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	err = c.checkTx(ctx, tx.Hash(), "mint")
	if err != nil {
		c.logger.Error(err)
		return 0, err
	}

	nftStore := &database.NFTStore{
		TokenId: tokenId.Uint64(),
		Address: to.Hex(),
		Cid:     cid,
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

func (c *NFTContract) checkTx(ctx context.Context, txHash common.Hash, name string) error {
	var receipt *etypes.Receipt

	t := checkTxSleepTime
	for i := 0; i < 10; i++ {
		// select {
		// case <-ctx.Done():
		// 	return nil
		// case <-time.After(time.Duration(t) * time.Second):
		// }
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
