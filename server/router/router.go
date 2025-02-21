package router

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	klog "github.com/go-kratos/kratos/v2/log"

	auth "github.com/memoio/xspace-server/authentication"
	"github.com/memoio/xspace-server/contract/nft"
)

type handler struct {
	context context.Context
	logger  *klog.Helper
	// store
	authController *auth.AuthController
	nftController  *nft.NFTController
}

func NewRouter(ctx context.Context, chain string, sk string, r *gin.RouterGroup) error {
	logger := klog.With(klog.NewStdLogger(os.Stdout),
		"ts", klog.DefaultTimestamp,
		"caller", klog.DefaultCaller,
	)

	loggers := klog.NewHelper(logger)

	authController, err := auth.NewAuthController(sk)
	if err != nil {
		return err
	}

	nftController, err := nft.NewNFTController(
		common.HexToAddress("0xa75150D716423c069529A3B2908Eb454e0a00Dfc"),
		"https://devchain.metamemo.one:8501",
		sk,
		loggers)
	if err != nil {
		return err
	}

	h := &handler{
		context:        ctx,
		nftController:  nftController,
		authController: authController,
		logger:         loggers,
	}

	LoadNFTModule(r.Group("/nft"), h)
	// LoadReferModule(r.Group("/refer"), h)
	LoadPointModules(r.Group("/"), h)
	LoadAuthModule(r.Group("/"), h)
	return nil
}
