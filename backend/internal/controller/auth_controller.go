package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/middleware"
	"github.com/enricojoe/dgram/backend/internal/service"
	"github.com/enricojoe/dgram/backend/internal/util"
)

// AuthController exposes registration, login, refresh and the current-user
// endpoint.
type AuthController struct {
	auth *service.AuthService
}

func NewAuthController(auth *service.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

// RegisterPublic mounts the unauthenticated auth routes.
func (a *AuthController) RegisterPublic(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/register", a.register)
	auth.POST("/login", a.login)
	auth.POST("/refresh", a.refresh)
}

// RegisterProtected mounts auth routes that require a valid access token.
func (a *AuthController) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/me", a.me)
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type authResponse struct {
	User         any    `json:"user"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (a *AuthController) register(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if !validCredentials(c, req) {
		return
	}

	user, tokens, err := a.auth.Register(strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			util.Fail(c, http.StatusConflict, "email already registered")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not register user")
		return
	}

	util.Created(c, authResponse{
		User:         user.View(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (a *AuthController) login(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if !validCredentials(c, req) {
		return
	}

	user, tokens, err := a.auth.Login(strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			util.Fail(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not log in")
		return
	}

	util.OK(c, authResponse{
		User:         user.View(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (a *AuthController) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		util.Fail(c, http.StatusBadRequest, "refreshToken must not be empty")
		return
	}

	tokens, err := a.auth.Refresh(req.RefreshToken)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	util.OK(c, tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (a *AuthController) me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := a.auth.GetUser(userID)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	util.OK(c, user.View())
}

// validCredentials performs basic input validation, writing a 400 and
// returning false when the email/password are missing.
func validCredentials(c *gin.Context, req credentialsRequest) bool {
	if strings.TrimSpace(req.Email) == "" {
		util.Fail(c, http.StatusBadRequest, "email must not be empty")
		return false
	}
	if req.Password == "" {
		util.Fail(c, http.StatusBadRequest, "password must not be empty")
		return false
	}
	return true
}
