package router

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

type handle struct {
	logger *log.Helper
}

func NewRouter(r *gin.Engine) {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)
	loggers := log.NewHelper(logger)

	h := &handle{
		logger: loggers,
	}
	loadApiMoudles(r.Group("/api"), h)

}

func loadApiMoudles(r *gin.RouterGroup, h *handle) {
	loadSpaceMoudles(r.Group("/space"), h)
	loadEarningMoudles(r.Group("/earning"), h)
	loadRankMoudles(r.Group("/rank"), h)
	loadTasksMoudles(r.Group("/tasks"), h)
	loadReferMoudles(r.Group("/refer"), h)
}
