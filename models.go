package main

import "time"

// User 用户
type User struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username       string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Nickname       string    `json:"nickname" gorm:"size:50"`
	HashedPassword string    `json:"-" gorm:"not null;size:128"`
	Role           string    `json:"role" gorm:"default:'user';size:20"`
	Status         int       `json:"status" gorm:"default:1"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Role 角色
type Role struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;size:20"`
	Name        string    `json:"name" gorm:"not null;size:50"`
	Description string    `json:"description" gorm:"size:200"`
	Status      int       `json:"status" gorm:"default:1"`
}

// Permission 权限
type Permission struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string `json:"code" gorm:"uniqueIndex;not null;size:50"`
	Name        string `json:"name" gorm:"not null;size:50"`
	Description string `json:"description" gorm:"size:200"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           int `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       int `json:"role_id" gorm:"not null"`
	PermissionID int `json:"permission_id" gorm:"not null"`
}

// UserPermission 用户权限关联
type UserPermission struct {
	ID           int `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int `json:"user_id" gorm:"not null"`
	PermissionID int `json:"permission_id" gorm:"not null"`
}

// LandContract 土地承包合同
type LandContract struct {
	ID                     int        `json:"id" gorm:"primaryKey;autoIncrement"`
	ContractCode           string     `json:"contract_code" gorm:"uniqueIndex;not null;size:32"`
	ContractorName         string     `json:"contractor_name" gorm:"not null;size:50"`
	ContractTotalArea1994  float64    `json:"contract_total_area_1994" gorm:"type:decimal(10,4);not null"`
	ConfirmedTotalArea     float64    `json:"confirmed_total_area" gorm:"type:decimal(10,4);not null"`
	TotalPlots             int        `json:"total_plots" gorm:"not null"`
	ChangedArea            *float64   `json:"changed_area" gorm:"type:decimal(10,4)"`
	Remarks                *string    `json:"remarks" gorm:"type:text"`
	SignatureConfirmed     bool       `json:"signature_confirmed" gorm:"default:false"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	Plots                  []LandPlot `json:"plots" gorm:"foreignKey:ContractID;references:ID"`
}

// LandPlot 地块
type LandPlot struct {
	ID               int      `json:"id" gorm:"primaryKey;autoIncrement"`
	ContractID       int      `json:"contract_id" gorm:"not null"`
	PlotName         string   `json:"plot_name" gorm:"not null;size:100"`
	PlotCode         string   `json:"plot_code" gorm:"not null;size:20"`
	ContractArea1994 float64  `json:"contract_area_1994" gorm:"type:decimal(10,4);not null"`
	ConfirmedArea    float64  `json:"confirmed_area" gorm:"type:decimal(10,4);not null"`
	BoundaryEast     *string  `json:"boundary_east" gorm:"size:100"`
	BoundarySouth    *string  `json:"boundary_south" gorm:"size:100"`
	BoundaryWest     *string  `json:"boundary_west" gorm:"size:100"`
	BoundaryNorth    *string  `json:"boundary_north" gorm:"size:100"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
