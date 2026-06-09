package handler

import (
	"log"
	"net/http"
	"time"

	"marketing-platform/internal/middleware"
	"marketing-platform/internal/model"
	"marketing-platform/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userService: service.NewUserService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request parameters")
		return
	}

	user, err := h.userService.FindByUsername(req.Username)
	if err != nil {
		middleware.InternalError(c, "Database error")
		return
	}

	if user == nil {
		middleware.Unauthorized(c, "Invalid username or password")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		middleware.Unauthorized(c, "Invalid username or password")
		return
	}

	// 生成Token
	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		middleware.InternalError(c, "Failed to generate token")
		return
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	h.userService.Update(user)

	middleware.Success(c, gin.H{
		"token":      token,
		"user_id":    user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"last_login": user.LastLogin,
	})
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
	Nickname string `json:"nickname" binding:"omitempty,max=100"`
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// 检查用户名是否存在
	existingUser, _ := h.userService.FindByUsername(req.Username)
	if existingUser != nil {
		middleware.BadRequest(c, "Username already exists")
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.InternalError(c, "Failed to process password")
		return
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := h.userService.Create(user); err != nil {
		log.Printf("Register: failed to create user '%s': %v", req.Username, err)
		middleware.InternalError(c, "Failed to create user")
		return
	}

	middleware.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not found in context")
		return
	}

	user, err := h.userService.FindByID(userID.(uint))
	if err != nil || user == nil {
		middleware.NotFound(c, "User not found")
		return
	}

	middleware.Success(c, gin.H{
		"user_id":   user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"nickname":  user.Nickname,
		"status":    user.Status,
		"last_login": user.LastLogin,
		"created_at": user.CreatedAt,
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request parameters")
		return
	}

	userID, _ := c.Get("user_id")
	user, err := h.userService.FindByID(userID.(uint))
	if err != nil || user == nil {
		middleware.NotFound(c, "User not found")
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		middleware.BadRequest(c, "Incorrect old password")
		return
	}

	// 更新密码
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	if err := h.userService.Update(user); err != nil {
		middleware.InternalError(c, "Failed to update password")
		return
	}

	middleware.Success(c, nil)
}

// HealthCheck 健康检查
func (h *AuthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
