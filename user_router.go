package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserCreateReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

type UserUpdateReq struct {
	Username *string `json:"username"`
	Nickname *string `json:"nickname"`
	Role     *string `json:"role"`
	Status   *int    `json:"status"`
}

type UserAdminUpdateReq struct {
	Nickname *string `json:"nickname"`
	Role     *string `json:"role"`
	Status   *int    `json:"status"`
	Password *string `json:"password"`
}

func setupUserRouter(r *gin.Engine) {
	users := r.Group("/api/v1/users", authMiddleware(), requireAdmin())

	users.GET("", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if pageSize > 100 {
			pageSize = 100
		}
		keyword := c.Query("keyword")

		var total int64
		query := DB.Model(&User{})
		if keyword != "" {
			query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		query.Count(&total)

		var items []User
		query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items)

		c.JSON(200, gin.H{"total": total, "items": items})
	})

	users.POST("", func(c *gin.Context) {
		var req UserCreateReq
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
			Role:           req.Role,
			Status:         req.Status,
		}
		DB.Create(&user)
		c.JSON(201, user)
	})

	users.PUT("/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var user User
		if DB.First(&user, id).Error != nil {
			c.JSON(404, gin.H{"detail": "用户不存在"})
			return
		}

		var req UserAdminUpdateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		currentUser := getCurrentUser(c)
		if req.Role != nil && *req.Role != "admin" && user.ID == currentUser.ID {
			var otherAdmin User
			if DB.Where("role = ? AND id != ? AND status = 1", "admin", currentUser.ID).First(&otherAdmin).Error != nil {
				c.JSON(400, gin.H{"detail": "系统中至少需要保留一个管理员"})
				return
			}
		}

		updates := make(map[string]interface{})
		if req.Nickname != nil {
			updates["nickname"] = *req.Nickname
		}
		if req.Role != nil {
			updates["role"] = *req.Role
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}
		if req.Password != nil && *req.Password != "" {
			updates["hashed_password"], _ = hashPassword(*req.Password)
		}

		DB.Model(&user).Updates(updates)
		c.JSON(200, user)
	})

	users.DELETE("/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var user User
		if DB.First(&user, id).Error != nil {
			c.JSON(404, gin.H{"detail": "用户不存在"})
			return
		}

		currentUser := getCurrentUser(c)
		if user.ID == currentUser.ID {
			c.JSON(400, gin.H{"detail": "不能删除自己"})
			return
		}

		if user.Role == "admin" {
			var otherAdmin User
			if DB.Where("role = ? AND id != ? AND status = 1", "admin", user.ID).First(&otherAdmin).Error != nil {
				c.JSON(400, gin.H{"detail": "系统中至少需要保留一个管理员"})
				return
			}
		}

		user.Status = 0
		DB.Save(&user)
		c.Status(204)
	})

	users.POST("/:id/restore", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var user User
		if DB.First(&user, id).Error != nil {
			c.JSON(404, gin.H{"detail": "用户不存在"})
			return
		}

		user.Status = 1
		DB.Save(&user)
		c.JSON(200, user)
	})
}
