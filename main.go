package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func initData() {
	// 初始化权限
	defaultPermissions := []struct {
		Code string
		Name string
	}{
		{"user:create", "创建用户"},
		{"user:update", "编辑用户"},
		{"user:delete", "禁用用户"},
		{"user:read", "查看用户"},
		{"land:create", "创建土地确权"},
		{"land:update", "编辑土地确权"},
		{"land:delete", "删除土地确权"},
		{"land:read", "查看土地确权"},
	}

	permMap := make(map[string]int)
	for _, dp := range defaultPermissions {
		var perm Permission
		if DB.Where("code = ?", dp.Code).First(&perm).Error != nil {
			perm = Permission{Code: dp.Code, Name: dp.Name}
			DB.Create(&perm)
		}
		permMap[dp.Code] = perm.ID
	}

	// 初始化角色
	defaultRoles := []struct {
		Code string
		Name string
		Desc string
	}{
		{"admin", "管理员", "系统管理员，拥有所有权限"},
		{"user", "普通用户", "普通用户，默认仅查看权限"},
	}

	roleMap := make(map[string]int)
	for _, dr := range defaultRoles {
		var role Role
		if DB.Where("code = ?", dr.Code).First(&role).Error != nil {
			role = Role{Code: dr.Code, Name: dr.Name, Description: dr.Desc, Status: 1}
			DB.Create(&role)
		}
		roleMap[dr.Code] = role.ID
	}

	// admin 角色分配所有权限
	if adminRoleID, ok := roleMap["admin"]; ok {
		for _, permID := range permMap {
			var rp RolePermission
			if DB.Where("role_id = ? AND permission_id = ?", adminRoleID, permID).First(&rp).Error != nil {
				DB.Create(&RolePermission{RoleID: adminRoleID, PermissionID: permID})
			}
		}
	}

	// user 角色清除所有权限
	if userRoleID, ok := roleMap["user"]; ok {
		DB.Where("role_id = ?", userRoleID).Delete(&RolePermission{})
	}

	// 初始化 admin 用户
	var admin User
	if DB.Where("username = ?", "admin").First(&admin).Error != nil {
		hash, _ := hashPassword("admin")
		admin = User{
			Username:       "admin",
			Nickname:       "管理员",
			HashedPassword: hash,
			Role:           "admin",
			Status:         1,
		}
		DB.Create(&admin)
		log.Println("[初始化] 默认管理员用户 admin/admin 已创建")
	}

	log.Println("[初始化] 角色和权限数据已就绪")
}

func main() {
	InitDB()
	initData()

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
	}))

	// 路由
	setupAuthRouter(r)
	setupUserRouter(r)
	setupPermissionRouter(r)
	setupLandRouter(r)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        cfg.ProjectName,
			"version":     cfg.Version,
			"docs":        "/docs",
		})
	})

	log.Printf("[启动] %s 运行在 http://0.0.0.0:%s", cfg.ProjectName, cfg.Port)
	r.Run(":" + cfg.Port)
}
