package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func setupPermissionRouter(r *gin.Engine) {
	perms := r.Group("/api/v1/permissions", authMiddleware(), requireAdmin())

	perms.GET("", func(c *gin.Context) {
		var items []Permission
		DB.Order("id").Find(&items)

		result := make([]gin.H, 0, len(items))
		for _, p := range items {
			result = append(result, gin.H{
				"id":          p.ID,
				"code":        p.Code,
				"name":        p.Name,
				"description": p.Description,
			})
		}
		c.JSON(200, result)
	})

	perms.GET("/user/:user_id", func(c *gin.Context) {
		userID, _ := strconv.Atoi(c.Param("user_id"))

		var user User
		if DB.First(&user, userID).Error != nil {
			c.JSON(404, gin.H{"detail": "用户不存在"})
			return
		}

		var userPerms []UserPermission
		DB.Where("user_id = ?", userID).Find(&userPerms)

		permIDs := make([]int, 0, len(userPerms))
		for _, up := range userPerms {
			permIDs = append(permIDs, up.PermissionID)
		}

		c.JSON(200, gin.H{
			"user_id":       userID,
			"role":          user.Role,
			"permission_ids": permIDs,
		})
	})

	perms.PUT("/user/:user_id", func(c *gin.Context) {
		userID, _ := strconv.Atoi(c.Param("user_id"))

		var user User
		if DB.First(&user, userID).Error != nil {
			c.JSON(404, gin.H{"detail": "用户不存在"})
			return
		}

		if user.Role == "admin" {
			c.JSON(400, gin.H{"detail": "管理员用户的权限不可单独配置"})
			return
		}

		var req struct {
			PermissionIDs []int `json:"permission_ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		// 删除原有权限
		DB.Where("user_id = ?", userID).Delete(&UserPermission{})

		// 插入新权限
		for _, permID := range req.PermissionIDs {
			var perm Permission
			if DB.First(&perm, permID).Error == nil {
				DB.Create(&UserPermission{UserID: userID, PermissionID: permID})
			}
		}

		c.JSON(200, gin.H{"message": "权限更新成功"})
	})
}
