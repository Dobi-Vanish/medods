package errormsg

import "errors"

// ErrorResponse represents standard error response structure
// @name ErrorResponse.
type ErrorResponse struct {
	Error   bool   `json:"error" example:"true"`
	Message string `json:"message" example:"error description"`
}

var (
	ErrPasswordLength                = errors.New("password must be at least 8 characters long")
	ErrFetchUsers                    = errors.New("couldn't fetch all users")
	ErrUserNotExist                  = errors.New("user with this email does not exist")
	ErrInvalidPassword               = errors.New("invalid password")
	ErrFetchUser                     = errors.New("couldn't fetch user")
	ErrUserNotFound                  = errors.New("user does not exist")
	ErrScanUser                      = errors.New("failed to scan user")
	ErrInvalidID                     = errors.New("provided ID is invalid")
	ErrInvalidToken                  = errors.New("invalid access token from errormsg")
	ErrEmptyID                       = errors.New("empty ID parameter")
	ErrRepositoryError               = errors.New("repository error")
	ErrUnexpectedSigningMethod       = errors.New("unexpected signing method")
	ErrTokenValidation               = errors.New("token validation failed")
	ErrApplyMigrations               = errors.New("error during applying migrations")
	ErrConnectDB                     = errors.New("error during connecting to DB")
	ErrSetDialect                    = errors.New("error during setting dialect to postgres")
	ErrJSONDecode                    = errors.New("JSON decode has failed")
	ErrJSONMustContain               = errors.New("must contain at least one JSON value")
	ErrDSNRequired                   = errors.New("DSN is required")
	ErrServerPortRequired            = errors.New("server port is required")
	ErrPostgresConnectAttemptsFailed = errors.New("failed connect to Postgres after 10 attempts")
	ErrTokenExpired                  = errors.New("your auth has expired, please, authenticate again")
	ErrInvalidIP                     = errors.New("invalid IP while working with tokens")
	ErrInvalidRefreshToken           = errors.New("invalid refresh token from errormsg")
	ErrCompareHash                   = errors.New("error during comparing hash and sotre token")
	ErrStorage                       = errors.New("storage error")
	ErrUpdate                        = errors.New("failed to update")
)
