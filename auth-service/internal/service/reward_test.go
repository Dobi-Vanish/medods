package service_test

import (
	"auth-service/api/calltypes"
	"auth-service/internal/service"
	"auth-service/pkg/consts"
	"auth-service/pkg/errormsg"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

// MockRepository - мок репозитория для тестирования.
type MockRepository struct {
	mock.Mock
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func (m *MockRepository) Insert(user calltypes.User) (int, error) {
	args := m.Called(user)

	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetAll() ([]*calltypes.User, error) {
	args := m.Called()

	users, ok := args.Get(0).([]*calltypes.User)
	if !ok {
		return nil, fmt.Errorf("type assertion failed: expected []*calltypes.User, got %T", args.Get(0)) //nolint: err113
	}

	return users, args.Error(1) //nolint: wrapcheck
}

func (m *MockRepository) GetOne(id int) (*calltypes.User, error) {
	args := m.Called(id)

	user, ok := args.Get(0).(*calltypes.User)
	if !ok {
		return nil, fmt.Errorf("type assertion to *calltypes.User failed, got %T", args.Get(0)) //nolint: err113
	}

	return user, args.Error(1) //nolint: wrapcheck
}

func (m *MockRepository) GetByEmail(email string) (*calltypes.User, error) {
	args := m.Called(email)

	user, ok := args.Get(0).(*calltypes.User)
	if !ok {
		return nil, fmt.Errorf("type assertion to *calltypes.User failed, got %T", args.Get(0)) //nolint: err113
	}

	return user, args.Error(1) //nolint: wrapcheck
}

func (m *MockRepository) PasswordMatches(password string, user calltypes.User) (bool, error) {
	args := m.Called(password, user)

	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Update(user calltypes.User) error {
	args := m.Called(user)

	return args.Error(0) //nolint: wrapcheck
}

func (m *MockRepository) EmailCheck(email string) (*calltypes.User, error) {
	args := m.Called(email)

	user, ok := args.Get(0).(*calltypes.User)
	if !ok {
		return nil, fmt.Errorf("type assertion to *calltypes.User failed, got %T", args.Get(0)) //nolint: err113
	}

	return user, args.Error(1) //nolint: wrapcheck
}

func (m *MockRepository) StoreRefreshToken(userID int, hashedToken string) error {
	args := m.Called(userID, hashedToken)

	return args.Error(0) //nolint: wrapcheck
}

func (m *MockRepository) ValidateRefreshToken(rawToken string, clientIP string, id int) (bool, error) {
	args := m.Called(rawToken, clientIP, id)

	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) UpdateRefreshToken(id int, rawToken string) error {
	args := m.Called(id, rawToken)

	return args.Error(0) //nolint: wrapcheck
}

func TestRewardService_Registrate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockRepository)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Successful registration",
			requestBody: `{
				"email": "test@example.com",
				"firstName": "Test",
				"lastName": "User",
				"password": "securepassword123"
			}`,
			mockSetup: func(m *MockRepository) {
				m.On("Insert", mock.AnythingOfType("calltypes.User")).Return(1, nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedError:  false,
		},
		{
			name: "Short password",
			requestBody: `{
				"email": "test@example.com",
				"firstName": "Test",
				"lastName": "User",
				"password": "short"
			}`,
			mockSetup:      func(_ *MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Repository error",
			requestBody: `{
				"email": "test@example.com",
				"firstName": "Test",
				"lastName": "User",
				"password": "securepassword123"
			}`,
			mockSetup: func(m *MockRepository) {
				m.On("Insert", mock.AnythingOfType("calltypes.User")).Return(0, errormsg.ErrRepositoryError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewRewardService(mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/registrate", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			svc.Registrate(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.mockSetup != nil {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestRewardService_Authenticate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockRepository)
		expectedStatus int
	}{
		{
			name: "Successful authentication",
			requestBody: `{
                "email": "test@example.com",
                "password": "correctpassword"
            }`,
			mockSetup: func(m *MockRepository) { //nolint:varnamelen
				user := &calltypes.User{
					ID:        1,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Password:  "hashedpassword",
				}
				m.On("GetByEmail", "test@example.com").Return(user, nil)
				m.On("PasswordMatches", "correctpassword", *user).Return(true, nil)
				m.On("StoreRefreshToken", user.ID, mock.AnythingOfType("string")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid credentials",
			requestBody: `{
                "email": "test@example.com",
                "password": "wrongpassword"
            }`,
			mockSetup: func(m *MockRepository) { //nolint:varnamelen
				user := &calltypes.User{
					ID:        1,
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Password:  "hashedpassword",
				}
				m.On("GetByEmail", "test@example.com").Return(user, nil)
				m.On("PasswordMatches", "wrongpassword", *user).Return(false, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User not found",
			requestBody: `{
                "email": "nonexistent@example.com",
                "password": "somepassword"
            }`,
			mockSetup: func(m *MockRepository) {
				m.On("GetByEmail", "nonexistent@example.com").Return((*calltypes.User)(nil), errormsg.ErrUserNotExist)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewRewardService(mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/authenticate", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			svc.Authenticate(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				cookies := rr.Result().Cookies()
				assert.Len(t, cookies, 2)
				assert.Equal(t, "accessToken", cookies[0].Name)
				assert.Equal(t, "refreshToken", cookies[1].Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRewardService_GetLeaderboard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*MockRepository)
		expectedStatus int
	}{
		{
			name: "Successful fetch",
			mockSetup: func(m *MockRepository) {
				users := []*calltypes.User{
					{ID: 1, FirstName: "User1"},
					{ID: 2, FirstName: "User2"},
				}
				m.On("GetAll").Return(users, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Repository error",
			mockSetup: func(m *MockRepository) {
				m.On("GetAll").Return([]*calltypes.User{}, errormsg.ErrRepositoryError)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewRewardService(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/users/leaderboard", nil)

			rr := httptest.NewRecorder()

			svc.GetLeaderboard(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRewardService_RetrieveOne(t *testing.T) {
	t.Parallel()

	testUser := &calltypes.User{
		ID:        123,
		FirstName: "test",
	}

	tests := []struct {
		name          string
		urlID         string
		repoResponse  *calltypes.User
		repoError     error
		expectedCode  int
		expectedError bool
	}{
		{
			name:          "successful retrieval",
			urlID:         "123",
			repoResponse:  testUser,
			repoError:     nil,
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "invalid ID in URL",
			urlID:         "abc",
			repoResponse:  nil,
			repoError:     nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name:          "user not found",
			urlID:         "123",
			repoResponse:  nil,
			repoError:     errormsg.ErrUserNotFound,
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			if tt.urlID == "123" {
				mockRepo.On("GetOne", 123).Return(tt.repoResponse, tt.repoError)
			}

			svc := &service.RewardService{Repo: mockRepo}

			req, err := http.NewRequest(http.MethodGet, "/users/"+tt.urlID+"/status", nil)
			require.NoError(t, err)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.urlID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := context.WithValue(req.Context(), userIDKey, "test-user")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			svc.RetrieveOne(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if !tt.expectedError && tt.expectedCode == http.StatusOK {
				var response *calltypes.JSONResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Error)
				assert.Equal(t, "Retrieved one user from the database", response.Message)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRewardService_Provide(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		urlID           string
		ip              string
		storeTokenError error
		expectedCode    int
		expectedError   bool
	}{
		{
			name:            "successful token provision",
			urlID:           "123",
			ip:              "192.168.1.1",
			storeTokenError: nil,
			expectedCode:    http.StatusOK,
			expectedError:   false,
		},
		{
			name:            "invalid ID in URL",
			urlID:           "abc",
			ip:              "192.168.1.1",
			storeTokenError: nil,
			expectedCode:    http.StatusBadRequest,
			expectedError:   true,
		},
		{
			name:            "empty IP address",
			urlID:           "123",
			ip:              "",
			storeTokenError: nil,
			expectedCode:    http.StatusBadRequest,
			expectedError:   true,
		},
		{
			name:            "failed to store refresh token",
			urlID:           "123",
			ip:              "192.168.1.1",
			storeTokenError: errormsg.ErrStorage,
			expectedCode:    http.StatusInternalServerError,
			expectedError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			if tc.urlID == "123" && tc.ip != "" {
				mockRepo.On("StoreRefreshToken", 123, mock.AnythingOfType("string")).Return(tc.storeTokenError)
			}

			svc := &service.RewardService{Repo: mockRepo}

			req, err := http.NewRequest(http.MethodPost, "/users/"+tc.urlID+"/tokens", nil)
			require.NoError(t, err)

			if tc.ip != "" {
				req.RemoteAddr = tc.ip + ":12345"
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.urlID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			svc.Provide(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if !tc.expectedError {
				cookies := rr.Result().Cookies()
				assert.Len(t, cookies, 2)

				var accessCookie, refreshCookie *http.Cookie

				for _, cookie := range cookies {
					switch cookie.Name {
					case "accessToken":
						accessCookie = cookie
					case "refreshToken":
						refreshCookie = cookie
					}
				}

				assert.NotNil(t, accessCookie)
				assert.NotEmpty(t, accessCookie.Value)
				assert.WithinDuration(t, time.Now().Add(consts.AccessTokenExpireTime), accessCookie.Expires, time.Second)

				assert.NotNil(t, refreshCookie)
				assert.NotEmpty(t, refreshCookie.Value)
				assert.WithinDuration(t, time.Now().Add(consts.RefreshTokenExpireTime), refreshCookie.Expires, time.Second)

				var response calltypes.JSONResponse
				err = json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Error)
				assert.Equal(t, "Tokens has been successfully provided", response.Message)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRewardService_Refresh(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		urlID              string
		ip                 string
		cookieValue        string
		validationResult   bool
		validationError    error
		updateTokenError   error
		expectedCode       int
		expectTokenRefresh bool
	}{
		{
			name:               "successful token refresh",
			urlID:              "123",
			ip:                 "192.168.1.1",
			cookieValue:        "valid_refresh_token",
			validationResult:   true,
			validationError:    nil,
			updateTokenError:   nil,
			expectedCode:       http.StatusOK,
			expectTokenRefresh: true,
		},
		{
			name:               "invalid ID in URL",
			urlID:              "abc",
			ip:                 "192.168.1.1",
			cookieValue:        "valid_refresh_token",
			expectedCode:       http.StatusBadRequest,
			expectTokenRefresh: false,
		},
		{
			name:               "missing refresh token cookie",
			urlID:              "123",
			ip:                 "192.168.1.1",
			cookieValue:        "",
			expectedCode:       http.StatusUnauthorized,
			expectTokenRefresh: false,
		},
		{
			name:               "invalid refresh token",
			urlID:              "123",
			ip:                 "192.168.1.1",
			cookieValue:        "invalid_refresh_token",
			validationResult:   false,
			validationError:    errormsg.ErrInvalidToken,
			expectedCode:       http.StatusUnauthorized,
			expectTokenRefresh: false,
		},
		{
			name:               "failed to update refresh token",
			urlID:              "123",
			ip:                 "192.168.1.1",
			cookieValue:        "valid_refresh_token",
			validationResult:   true,
			validationError:    nil,
			updateTokenError:   errormsg.ErrUpdate,
			expectedCode:       http.StatusInternalServerError,
			expectTokenRefresh: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRepository)
			if tc.urlID == "123" && tc.cookieValue != "" {
				mockRepo.On("ValidateRefreshToken", tc.cookieValue, tc.ip, 123).Return(tc.validationResult, tc.validationError)

				if tc.validationResult && tc.validationError == nil {
					mockRepo.On("UpdateRefreshToken", 123, mock.AnythingOfType("string")).Return(tc.updateTokenError)
				}
			}

			svc := &service.RewardService{Repo: mockRepo}

			req, err := http.NewRequest(http.MethodPost, "/users/"+tc.urlID+"/refresh", nil)
			require.NoError(t, err)

			if tc.ip != "" {
				req.RemoteAddr = tc.ip + ":12345"
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.urlID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tc.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refreshToken",
					Value: tc.cookieValue,
				})
			}

			rr := httptest.NewRecorder()

			svc.Refresh(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectTokenRefresh {
				cookies := rr.Result().Cookies()
				assert.Len(t, cookies, 2)

				var accessCookie, refreshCookie *http.Cookie

				for _, cookie := range cookies {
					switch cookie.Name {
					case "accessToken":
						accessCookie = cookie
					case "refreshToken":
						refreshCookie = cookie
					}
				}

				assert.NotNil(t, accessCookie)
				assert.NotEmpty(t, accessCookie.Value)
				assert.WithinDuration(t, time.Now().Add(consts.AccessTokenExpireTime), accessCookie.Expires, time.Second)

				assert.NotNil(t, refreshCookie)
				assert.NotEmpty(t, refreshCookie.Value)
				assert.WithinDuration(t, time.Now().Add(consts.RefreshTokenExpireTime), refreshCookie.Expires, time.Second)

				var response calltypes.JSONResponse
				err = json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Error)
				assert.Equal(t, "Tokens has been successfully refreshed", response.Message)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
