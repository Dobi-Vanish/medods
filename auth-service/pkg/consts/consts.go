package consts

import "time"

const (
	DbTimeout              = time.Second * 3
	PassMinLength          = 8
	BcryptCost             = 12
	RefreshTokenExpireTime = 30 * 24 * time.Hour
	AtLeastPassLength      = 8
	AccessTokenExpireTime  = 15 * time.Minute
	RefreshTokenLength     = 32
	ConnectAttempts        = 10
	WaitBeforeAttempts     = 2
	MaxAge                 = 300
	Megabyte               = 1 << 20
	IdleTimeout            = 30
	WriteTimeout           = 10
	ReadTimeout            = 5
	TokenParts             = 2
	TestIP                 = "10.10.10.10"
)
