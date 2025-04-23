// Package service implements reward service API handlers.
package service

import (
	"auth-service/api/calltypes"
	"auth-service/api/server/httputils"
	"auth-service/internal/postgres/repository"
	"auth-service/internal/token"
	"auth-service/pkg/consts"
	"auth-service/pkg/errormsg"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NewRewardService(repo repository.Repository) *RewardService {
	return &RewardService{
		Repo:   repo,
		Client: &http.Client{},
	}
}

// GetIDFromURL godoc
// @Summary Extract ID from URL parameter
// @Description Parses and validates ID from URL path
// @Tags Utilities
// @Param paramName path string true "URL parameter name containing ID"
// @Success 200 {integer} int "Valid ID"
// @Failure 400 {object} calltypes.ErrorResponse "Invalid or empty ID"
// @Router /parse-id/{paramName} [get].
func GetIDFromURL(r *http.Request, paramName string) (int, error) {
	idStr := chi.URLParam(r, paramName)
	idStr = strings.TrimSpace(idStr)

	if idStr == "" {
		return 0, errormsg.ErrEmptyID
	}

	for _, c := range idStr {
		if c == '-' && len(idStr) > 1 {
			continue
		}

		if c < '0' || c > '9' {
			return 0, errormsg.ErrInvalidID
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errormsg.ErrInvalidID
	}

	return id, nil
}

func GetClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

// Registrate godoc
// @Summary Register new user
// @Description Creates new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param request body calltypes.RegisterRequest true "User registration data"
// @Success 202 {object} calltypes.JSONResponse
// @Failure 400 {object} calltypes.ErrorResponse "Invalid request data"
// @Router /register [post].
func (s *RewardService) Registrate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Password  string `json:"password"`
		Active    int    `json:"active,omitempty"`
		Score     int    `json:"score,omitempty"`
		Referrer  string `json:"referrer,omitempty"`
	}

	err := httputils.ReadJSON(w, r, &requestPayload)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}

	if len(requestPayload.Password) < consts.AtLeastPassLength {
		httputils.ErrorJSON(w, errormsg.ErrPasswordLength, http.StatusBadRequest)

		return
	}

	user := calltypes.User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    requestPayload.Active,
	}

	id, err := s.Repo.Insert(user)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Successfully created new user, id: %d", id),
	}

	err = httputils.WriteJSON(w, http.StatusAccepted, payload)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}

// GetLeaderboard godoc
// @Summary Get user leaderboard
// @Description Returns all users ordered by score
// @Tags Users
// @Produce json
// @Success 200 {object} calltypes.JSONResponse{data=[]calltypes.User}
// @Failure 400 {object} calltypes.ErrorResponse "Failed to fetch users"
// @Router /leaderboard [get].
func (s *RewardService) GetLeaderboard(w http.ResponseWriter, _ *http.Request) {
	users, err := s.Repo.GetAll()
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrFetchUsers, http.StatusBadRequest)

		return
	}

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: "Fetched all users",
		Data:    users,
	}

	err = httputils.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}

// Authenticate godoc
// @Summary Authenticate user
// @Description Logs in user and returns auth cookies
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body calltypes.LoginRequest true "Credentials"
// @Success 200 {object} calltypes.JSONResponse
// @Header 200 {string} Set-Cookie "accessToken"
// @Header 200 {string} Set-Cookie "refreshToken"
// @Failure 400 {object} calltypes.ErrorResponse "Invalid credentials"
// @Router /login [post].
func (s *RewardService) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httputils.ReadJSON(w, r, &requestPayload); err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}

	if requestPayload.Email == "" || requestPayload.Password == "" {
		httputils.ErrorJSON(w, errormsg.ErrUserNotFound, http.StatusBadRequest)

		return
	}

	user, err := s.Repo.GetByEmail(requestPayload.Email)
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrUserNotExist, http.StatusBadRequest)

		return
	}

	valid, err := s.Repo.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		httputils.ErrorJSON(w, errormsg.ErrInvalidPassword, http.StatusBadRequest)

		return
	}

	tokenService := token.NewTokenService()

	ip := GetClientIP(r)
	if ip == "" {
		httputils.ErrorJSON(w, errormsg.ErrInvalidIP, http.StatusBadRequest)

		return
	}

	accessToken, hashedRefreshToken, err := tokenService.GenerateTokens(ip)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	err = s.Repo.StoreRefreshToken(user.ID, hashedRefreshToken)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.AccessTokenExpireTime),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    hashedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.RefreshTokenExpireTime),
	})

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Welcome back, %s!", user.FirstName),
		Data:    map[string]interface{}{"user_id": user.ID},
	}

	err = httputils.WriteJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}

// Provide godoc
// @Summary Provide new tokens
// @Description Generates and returns new access and refresh tokens for user
// @Tags Auth
// @Param id path int true "User ID"
// @Produce json
// @Success 200 {object} calltypes.JSONResponse
// @Header 200 {string} Set-Cookie "accessToken"
// @Header 200 {string} Set-Cookie "refreshToken"
// @Failure 400 {object} calltypes.ErrorResponse "Invalid ID or IP"
// @Failure 500 {object} calltypes.ErrorResponse "Internal server error"
// @Router /users/{id}/tokens [post].
func (s *RewardService) Provide(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDFromURL(r, "id")
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrInvalidID, http.StatusBadRequest)

		return
	}

	tokenService := token.NewTokenService()

	ip := GetClientIP(r)
	if ip == "" {
		httputils.ErrorJSON(w, errormsg.ErrInvalidIP, http.StatusBadRequest)

		return
	}

	accessToken, hashedRefreshToken, err := tokenService.GenerateTokens(ip)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	err = s.Repo.StoreRefreshToken(id, hashedRefreshToken)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.AccessTokenExpireTime),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    hashedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.RefreshTokenExpireTime),
	})

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: "Tokens has been successfully provided",
	}

	err = httputils.WriteJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Refreshes access and refresh tokens using valid refresh token
// @Tags Auth
// @Param id path int true "User ID"
// @Produce json
// @Success 200 {object} calltypes.JSONResponse
// @Header 200 {string} Set-Cookie "accessToken"
// @Header 200 {string} Set-Cookie "refreshToken"
// @Failure 400 {object} calltypes.ErrorResponse "Invalid ID"
// @Failure 401 {object} calltypes.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} calltypes.ErrorResponse "Internal server error"
// @Router /users/{id}/refresh [post].
func (s *RewardService) Refresh(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDFromURL(r, "id")
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrInvalidID, http.StatusBadRequest)

		return
	}

	refreshCookie, err := r.Cookie("refreshToken")
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusUnauthorized)

		return
	}

	ip := GetClientIP(r)

	ok, err := s.Repo.ValidateRefreshToken(refreshCookie.Value, ip, id)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusUnauthorized)

		return
	}

	if !ok {
		httputils.ErrorJSON(w, err, http.StatusUnauthorized)

		return
	}

	tokenService := token.NewTokenService()

	accessToken, hashedRefreshToken, err := tokenService.GenerateTokens(ip)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	err = s.Repo.UpdateRefreshToken(id, hashedRefreshToken)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.AccessTokenExpireTime),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    hashedRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(consts.RefreshTokenExpireTime),
	})

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: "Tokens has been successfully refreshed",
	}

	err = httputils.WriteJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}

// RetrieveOne godoc
// @Summary Get user by ID
// @Description Returns single user data
// @Tags Users
// @Param id path int true "User ID"
// @Produce json
// @Success 200 {object} calltypes.JSONResponse{data=calltypes.User}
// @Failure 400 {object} calltypes.ErrorResponse "User not found"
// @Router /users/{id} [get].
func (s *RewardService) RetrieveOne(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDFromURL(r, "id")
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrInvalidID, http.StatusBadRequest)

		return
	}

	user, err := s.Repo.GetOne(id)
	if err != nil {
		httputils.ErrorJSON(w, errormsg.ErrFetchUser, http.StatusBadRequest)

		return
	}

	payload := calltypes.JSONResponse{
		Error:   false,
		Message: "Retrieved one user from the database",
		Data:    user,
	}

	err = httputils.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		httputils.ErrorJSON(w, err, http.StatusBadRequest)

		return
	}
}
