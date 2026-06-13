package handlers

import (
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"

	"kenaito-notebook-server/services"
)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type AuthHandler struct {
	db      *services.Database
	auth    *services.AuthService
	email   *services.EmailService
	verify  *services.VerificationService
}

func NewAuthHandler(db *services.Database, auth *services.AuthService, email *services.EmailService, verify *services.VerificationService) *AuthHandler {
	return &AuthHandler{db: db, auth: auth, email: email, verify: verify}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SendCodeRequest struct {
	Email   string `json:"email" binding:"required"`
	Purpose string `json:"purpose" binding:"required"`
}

type VerifyCodeRequest struct {
	Email   string `json:"email" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Purpose string `json:"purpose" binding:"required"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// SendCode 发送验证码
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Auth] SendCode bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整信息"})
		return
	}

	if !isValidEmail(req.Email) {
		log.Printf("[Auth] SendCode invalid email: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	if req.Purpose != "register" && req.Purpose != "reset" {
		log.Printf("[Auth] SendCode invalid purpose: %s", req.Purpose)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求类型"})
		return
	}

	// 注册时检查邮箱是否已注册
	if req.Purpose == "register" {
		existing, _ := h.db.GetUserByEmail(req.Email)
		if existing != nil {
			log.Printf("[Auth] SendCode email already registered: %s", req.Email)
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
			return
		}
	}

	// 重置密码时检查邮箱是否存在
	if req.Purpose == "reset" {
		existing, _ := h.db.GetUserByEmail(req.Email)
		if existing == nil {
			log.Printf("[Auth] SendCode email not found for reset: %s", req.Email)
			c.JSON(http.StatusNotFound, gin.H{"error": "邮箱未注册"})
			return
		}
	}

	// 限频检查
	if !h.verify.CheckRateLimit(req.Email, req.Purpose) {
		log.Printf("[Auth] SendCode rate limit exceeded: %s, purpose: %s", req.Email, req.Purpose)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "发送过于频繁，请 60 秒后重试"})
		return
	}

	code := h.verify.GenerateCode(req.Email, req.Purpose)
	if err := h.email.SendVerificationCode(req.Email, code, req.Purpose); err != nil {
		log.Printf("[Auth] SendCode send email failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

// VerifyCode 验证验证码
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Auth] VerifyCode bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整信息"})
		return
	}

	if err := h.verify.VerifyCode(req.Email, req.Code, req.Purpose); err != nil {
		log.Printf("[Auth] VerifyCode failed: email=%s, purpose=%s, err=%v", req.Email, req.Purpose, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码正确"})
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Auth] Register bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整信息"})
		return
	}

	if !isValidEmail(req.Email) {
		log.Printf("[Auth] Register invalid email: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	// 检查邮箱是否已注册
	existing, _ := h.db.GetUserByEmail(req.Email)
	if existing != nil {
		log.Printf("[Auth] Register email already registered: %s", req.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 验证并消耗验证码
	if err := h.verify.UseCode(req.Email, req.Code, "register"); err != nil {
		log.Printf("[Auth] Register use code failed: email=%s, err=%v", req.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 哈希密码
	hashedPassword, err := h.auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("[Auth] Register hash password failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	// 创建用户
	user, err := h.db.CreateUser(req.Email, hashedPassword)
	if err != nil {
		log.Printf("[Auth] Register create user failed: email=%s, err=%v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	// 生成 token
	token, err := h.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("[Auth] Register generate token failed: userID=%d, err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Auth] Login bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写邮箱和密码"})
		return
	}

	if !isValidEmail(req.Email) {
		log.Printf("[Auth] Login invalid email: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	// 查找用户
	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("[Auth] Login user not found: %s, err=%v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}

	// 验证密码
	if !h.auth.CheckPassword(req.Password, user.HashedPassword) {
		log.Printf("[Auth] Login password mismatch: email=%s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}

	// 生成 token
	token, err := h.auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("[Auth] Login generate token failed: userID=%d, err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Auth] ResetPassword bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整信息"})
		return
	}

	if !isValidEmail(req.Email) {
		log.Printf("[Auth] ResetPassword invalid email: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	// 验证并消耗验证码
	if err := h.verify.UseCode(req.Email, req.Code, "reset"); err != nil {
		log.Printf("[Auth] ResetPassword use code failed: email=%s, err=%v", req.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("[Auth] ResetPassword user not found: %s, err=%v", req.Email, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 哈希新密码
	hashedPassword, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("[Auth] ResetPassword hash password failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置失败"})
		return
	}

	// 更新密码
	if err := h.db.UpdatePassword(user.ID, hashedPassword); err != nil {
		log.Printf("[Auth] ResetPassword update password failed: userID=%d, err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}

// GetMe 获取当前用户信息
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("[Auth] GetMe userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	user, err := h.db.GetUserByID(userID.(int64))
	if err != nil {
		log.Printf("[Auth] GetMe user not found: userID=%v, err=%v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[Auth] Middleware missing Authorization header: ip=%s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 提取 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[Auth] Middleware invalid auth format: ip=%s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证格式"})
			c.Abort()
			return
		}

		payload, err := auth.ValidateToken(parts[1])
		if err != nil {
			log.Printf("[Auth] Middleware token validation failed: ip=%s, err=%v", c.ClientIP(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", payload.UserID)
		c.Set("email", payload.Email)
		c.Next()
	}
}
