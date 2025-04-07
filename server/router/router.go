package router

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	klog "github.com/go-kratos/kratos/v2/log"

	auth "github.com/memoio/xspace-server/authentication"
	"github.com/memoio/xspace-server/config"
	"github.com/memoio/xspace-server/nft"
	"github.com/memoio/xspace-server/point"
)

type Router struct {
	handler *handler
}

type handler struct {
	context context.Context
	logger  *klog.Helper
	// store
	authController  *auth.AuthController
	nftController   *nft.NFTController
	pointController *point.PointController
}

func NewRouter(ctx context.Context, chain string, sk string, r *gin.RouterGroup) (*Router, error) {
	logger := klog.With(klog.NewStdLogger(os.Stdout),
		"ts", klog.DefaultTimestamp,
		"caller", klog.DefaultCaller,
	)

	loggers := klog.NewHelper(logger)

	authController, err := auth.NewAuthController(sk)
	if err != nil {
		loggers.Error(err)
		return nil, err
	}

	endpoint, nftAddr := config.GetContractInfoByChain(chain)
	nftController, err := nft.NewNFTController(nftAddr, endpoint, sk, loggers)
	if err != nil {
		loggers.Error(err)
		return nil, err
	}

	pointController, err := point.NewPointController()
	if err != nil {
		loggers.Error(err)
		return nil, err
	}

	h := &handler{
		context:         ctx,
		nftController:   nftController,
		authController:  authController,
		pointController: pointController,
		logger:          loggers,
	}

	LoadNFTModule(r.Group("/nft"), h)
	// LoadReferModule(r.Group("/refer"), h)
	LoadPointModules(r.Group("/"), h)
	LoadAuthModule(r.Group("/"), h)
	return &Router{handler: h}, nil
}

func (r *Router) Start(ctx context.Context) {
	r.handler.nftController.Start(ctx)
}

func (r *Router) Stop() error {
	return r.handler.nftController.Stop()
}
