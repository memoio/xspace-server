package router

import (
	"os"

	"github.com/gin-gonic/gin"
	klog "github.com/go-kratos/kratos/v2/log"

	// "github.com/memoio/xspace-server/auth"
	auth "github.com/memoio/xspace-server/authentication"
	"github.com/memoio/xspace-server/contract/nft"
)

type handler struct {
	logger *klog.Helper
	// store
	authController *auth.AuthController
	nftController  *nft.NFTController
}

func NewRouter(chain string, r *gin.RouterGroup) error {
	logger := klog.With(klog.NewStdLogger(os.Stdout),
		"ts", klog.DefaultTimestamp,
		"caller", klog.DefaultCaller,
	)

	loggers := klog.NewHelper(logger)

	nftController, err := nft.NewNFTController()
	if err != nil {
		return err
	}

	h := &handler{
		nftController: nftController,
		logger:        loggers,
	}

	LoadNFTModule(r.Group("/nft"), h)
	LoadReferModule(r.Group("/refer"), h)
	LoadPointModules(r.Group("/"), h)
	LoadAuthModule(r.Group("/"), h)
	return nil
}
