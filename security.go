package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ======== bcrypt ========

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ======== JWT ========

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * time.Duration(cfg.JWTHours)).Unix(),
	})
	return token.SignedString([]byte(cfg.JWTSecret))
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ======== Auth Middleware ========

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"detail": "未提供认证信息"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := parseToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"detail": "认证失败"})
			c.Abort()
			return
		}

		username := claims["sub"].(string)
		var user User
		if result := DB.Where("username = ?", username).First(&user); result.Error != nil {
			c.JSON(401, gin.H{"detail": "用户不存在"})
			c.Abort()
			return
		}

		if user.Status != 1 {
			c.JSON(403, gin.H{"detail": "用户已被禁用"})
			c.Abort()
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

func getCurrentUser(c *gin.Context) User {
	return c.MustGet("currentUser").(User)
}

// ======== Permission Check ========

func getUserPermissions(userID int, roleCode string) []string {
	if roleCode == "admin" {
		return []string{"*"}
	}

	permSet := make(map[string]bool)

	// 角色权限
	var role Role
	if result := DB.Where("code = ? AND status = 1", roleCode).First(&role); result.Error == nil {
		var rolePerms []RolePermission
		DB.Where("role_id = ?", role.ID).Find(&rolePerms)
		for _, rp := range rolePerms {
			var perm Permission
			if DB.First(&perm, rp.PermissionID).Error == nil {
				permSet[perm.Code] = true
			}
		}
	}

	// 用户直接权限
	var userPerms []UserPermission
	DB.Where("user_id = ?", userID).Find(&userPerms)
	for _, up := range userPerms {
		var perm Permission
		if DB.First(&perm, up.PermissionID).Error == nil {
			permSet[perm.Code] = true
		}
	}

	perms := make([]string, 0, len(permSet))
	for code := range permSet {
		perms = append(perms, code)
	}
	return perms
}

func requirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getCurrentUser(c)
		perms := getUserPermissions(user.ID, user.Role)

		for _, p := range perms {
			if p == "*" || p == permCode {
				c.Next()
				return
			}
		}

		c.JSON(403, gin.H{"detail": "权限不足，需要权限: " + permCode})
		c.Abort()
	}
}

func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getCurrentUser(c)
		if user.Role != "admin" {
			c.JSON(403, gin.H{"detail": "权限不足，需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
