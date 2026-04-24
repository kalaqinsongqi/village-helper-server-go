package main

import (
	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname"`
}

func setupAuthRouter(r *gin.Engine) {
	auth := r.Group("/api/v1/auth")

	auth.POST("/register", func(c *gin.Context) {
		var req RegisterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		var existing User
		if DB.Where("username = ?", req.Username).First(&existing).Error == nil {
			c.JSON(409, gin.H{"detail": "用户名已存在"})
			return
		}

		hash, _ := hashPassword(req.Password)
		user := User{
			Username:       req.Username,
			Nickname:       req.Nickname,
			HashedPassword: hash,
			Role:           "user",
			Status:         1,
		}
		DB.Create(&user)
		c.JSON(201, user)
	})

	auth.POST("/login", func(c *gin.Context) {
		var req LoginReq
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		var user User
		if DB.Where("username = ?", req.Username).First(&user).Error != nil {
			c.JSON(401, gin.H{"detail": "用户名或密码错误"})
			return
		}

		if user.Status != 1 {
			c.JSON(403, gin.H{"detail": "用户已被禁用"})
			return
		}

		if !verifyPassword(req.Password, user.HashedPassword) {
			c.JSON(401, gin.H{"detail": "用户名或密码错误"})
			return
		}

		token, _ := generateToken(user.Username)
		perms := getUserPermissions(user.ID, user.Role)

		c.JSON(200, gin.H{
			"access_token": token,
			"token_type":   "bearer",
			"expires_in":   cfg.JWTHours * 3600,
			"permissions":  perms,
			"role":         user.Role,
		})
	})

	auth.GET("/me", authMiddleware(), func(c *gin.Context) {
		user := getCurrentUser(c)
		c.JSON(200, user)
	})
}
