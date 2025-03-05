package authstrategies

import (
	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type BaseStrategy struct {
	ApiCode string
	AuthURL     string
	Credentials models.Credentials
	Logger      logger.ILogger
	Cache       cache.Cacher
	Crypto      *crypto.Crypto
}
