package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

// AuthHandler exposes HTTP handlers related to authentication.
type AuthHandler struct {
	users *services.UserService
}

// NewAuthHandler constructs an AuthHandler instance.
func NewAuthHandler(users *services.UserService) *AuthHandler {
	return &AuthHandler{users: users}
}

func ensureCORSHeaders(c *gin.Context) {
	if origin := c.GetHeader("Origin"); origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", "Origin")
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var payload registerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		ensureCORSHeaders(c)
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos invalidos"})
		return
	}

	err := h.users.Register(c.Request.Context(), services.User{
		Email:    payload.Email,
		Password: payload.Password,
	})
	switch {
	case err == nil:
		ensureCORSHeaders(c)
		c.JSON(http.StatusCreated, gin.H{"message": "usuario registrado con exito"})
	case errors.Is(err, services.ErrInvalidUserInput):
		ensureCORSHeaders(c)
		c.JSON(http.StatusBadRequest, gin.H{"error": "email y clave son requeridos"})
	case errors.Is(err, services.ErrUserAlreadyExists):
		ensureCORSHeaders(c)
		c.JSON(http.StatusConflict, gin.H{"error": "usuario ya existe"})
	default:
		ensureCORSHeaders(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al registrar usuario"})
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user authentication.
func (h *AuthHandler) Login(c *gin.Context) {
	var payload loginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		ensureCORSHeaders(c)
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos invalidos"})
		return
	}

	err := h.users.Login(c.Request.Context(), payload.Email, payload.Password)
	switch {
	case err == nil:
		ensureCORSHeaders(c)
		c.JSON(http.StatusOK, gin.H{"message": "login exitoso"})
	case errors.Is(err, services.ErrInvalidCredentials):
		ensureCORSHeaders(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales invalidas"})
	default:
		ensureCORSHeaders(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al autenticar"})
	}
}

// ListUsers returns every registered user in its public form.
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.users.List(c.Request.Context())
	if err != nil {
		ensureCORSHeaders(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener usuarios"})
		return
	}
	ensureCORSHeaders(c)
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ClearUsers removes every user. Intended for testing scenarios.
func (h *AuthHandler) ClearUsers(c *gin.Context) {
	if err := h.users.Clear(c.Request.Context()); err != nil {
		ensureCORSHeaders(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al limpiar usuarios"})
		return
	}

	ensureCORSHeaders(c)
	c.JSON(http.StatusOK, gin.H{"message": "usuarios eliminados"})
}
