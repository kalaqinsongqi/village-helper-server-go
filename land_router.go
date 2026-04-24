package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type LandContractCreateReq struct {
	ContractCode          string           `json:"contract_code" binding:"required"`
	ContractorName        string           `json:"contractor_name" binding:"required"`
	ContractTotalArea1994 float64          `json:"contract_total_area_1994" binding:"required"`
	ConfirmedTotalArea    float64          `json:"confirmed_total_area" binding:"required"`
	TotalPlots            int              `json:"total_plots" binding:"required"`
	ChangedArea           *float64         `json:"changed_area"`
	Remarks               *string          `json:"remarks"`
	SignatureConfirmed    bool             `json:"signature_confirmed"`
	Plots                 []LandPlotCreateReq `json:"plots"`
}

type LandContractUpdateReq struct {
	ContractCode          *string  `json:"contract_code"`
	ContractorName        *string  `json:"contractor_name"`
	ContractTotalArea1994 *float64 `json:"contract_total_area_1994"`
	ConfirmedTotalArea    *float64 `json:"confirmed_total_area"`
	TotalPlots            *int     `json:"total_plots"`
	ChangedArea           *float64 `json:"changed_area"`
	Remarks               *string  `json:"remarks"`
	SignatureConfirmed    *bool    `json:"signature_confirmed"`
}

type LandPlotCreateReq struct {
	PlotName         string  `json:"plot_name" binding:"required"`
	PlotCode         string  `json:"plot_code" binding:"required"`
	ContractArea1994 float64 `json:"contract_area_1994" binding:"required"`
	ConfirmedArea    float64 `json:"confirmed_area" binding:"required"`
	BoundaryEast     *string `json:"boundary_east"`
	BoundarySouth    *string `json:"boundary_south"`
	BoundaryWest     *string `json:"boundary_west"`
	BoundaryNorth    *string `json:"boundary_north"`
}

type LandPlotUpdateReq struct {
	PlotName         *string `json:"plot_name"`
	PlotCode         *string `json:"plot_code"`
	ContractArea1994 *float64 `json:"contract_area_1994"`
	ConfirmedArea    *float64 `json:"confirmed_area"`
	BoundaryEast     *string `json:"boundary_east"`
	BoundarySouth    *string `json:"boundary_south"`
	BoundaryWest     *string `json:"boundary_west"`
	BoundaryNorth    *string `json:"boundary_north"`
}

func setupLandRouter(r *gin.Engine) {
	land := r.Group("/api/v1/land")

	// 合同相关
	land.POST("/contracts", authMiddleware(), requirePermission("land:create"), func(c *gin.Context) {
		var req LandContractCreateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		var existing LandContract
		if DB.Where("contract_code = ?", req.ContractCode).First(&existing).Error == nil {
			c.JSON(409, gin.H{"detail": "合同代码已存在"})
			return
		}

		contract := LandContract{
			ContractCode:          req.ContractCode,
			ContractorName:        req.ContractorName,
			ContractTotalArea1994: req.ContractTotalArea1994,
			ConfirmedTotalArea:    req.ConfirmedTotalArea,
			TotalPlots:            len(req.Plots),
			ChangedArea:           req.ChangedArea,
			Remarks:               req.Remarks,
			SignatureConfirmed:    req.SignatureConfirmed,
		}
		DB.Create(&contract)

		for _, plotReq := range req.Plots {
			DB.Create(&LandPlot{
				ContractID:       contract.ID,
				PlotName:         plotReq.PlotName,
				PlotCode:         plotReq.PlotCode,
				ContractArea1994: plotReq.ContractArea1994,
				ConfirmedArea:    plotReq.ConfirmedArea,
				BoundaryEast:     plotReq.BoundaryEast,
				BoundarySouth:    plotReq.BoundarySouth,
				BoundaryWest:     plotReq.BoundaryWest,
				BoundaryNorth:    plotReq.BoundaryNorth,
			})
		}

		DB.Preload("Plots").First(&contract, contract.ID)
		c.JSON(201, contract)
	})

	land.GET("/contracts", authMiddleware(), requirePermission("land:read"), func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if pageSize > 100 {
			pageSize = 100
		}
		keyword := c.Query("keyword")
		contractorName := c.Query("contractor_name")
		contractCode := c.Query("contract_code")
		signatureConfirmed := c.Query("signature_confirmed")

		var total int64
		query := DB.Model(&LandContract{})
		if keyword != "" {
			query = query.Where("contract_code LIKE ? OR contractor_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
		if contractorName != "" {
			query = query.Where("contractor_name LIKE ?", "%"+contractorName+"%")
		}
		if contractCode != "" {
			query = query.Where("contract_code LIKE ?", "%"+contractCode+"%")
		}
		if signatureConfirmed != "" {
			if signatureConfirmed == "true" {
				query = query.Where("signature_confirmed = ?", true)
			} else if signatureConfirmed == "false" {
				query = query.Where("signature_confirmed = ?", false)
			}
		}
		query.Count(&total)

		var items []LandContract
		query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items)

		c.JSON(200, gin.H{"total": total, "items": items})
	})

	land.GET("/contracts/:id", authMiddleware(), requirePermission("land:read"), func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var contract LandContract
		if DB.Preload("Plots").First(&contract, id).Error != nil {
			c.JSON(404, gin.H{"detail": "记录不存在"})
			return
		}
		c.JSON(200, contract)
	})

	land.PUT("/contracts/:id", authMiddleware(), requirePermission("land:update"), func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var contract LandContract
		if DB.First(&contract, id).Error != nil {
			c.JSON(404, gin.H{"detail": "记录不存在"})
			return
		}

		var req LandContractUpdateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		updates := make(map[string]interface{})
		if req.ContractCode != nil {
			updates["contract_code"] = *req.ContractCode
		}
		if req.ContractorName != nil {
			updates["contractor_name"] = *req.ContractorName
		}
		if req.ContractTotalArea1994 != nil {
			updates["contract_total_area_1994"] = *req.ContractTotalArea1994
		}
		if req.ConfirmedTotalArea != nil {
			updates["confirmed_total_area"] = *req.ConfirmedTotalArea
		}
		if req.TotalPlots != nil {
			updates["total_plots"] = *req.TotalPlots
		}
		if req.ChangedArea != nil {
			updates["changed_area"] = *req.ChangedArea
		}
		if req.Remarks != nil {
			updates["remarks"] = *req.Remarks
		}
		if req.SignatureConfirmed != nil {
			updates["signature_confirmed"] = *req.SignatureConfirmed
		}

		DB.Model(&contract).Updates(updates)
		c.JSON(200, contract)
	})

	land.DELETE("/contracts/:id", authMiddleware(), requirePermission("land:delete"), func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var contract LandContract
		if DB.First(&contract, id).Error != nil {
			c.JSON(404, gin.H{"detail": "记录不存在"})
			return
		}
		DB.Delete(&contract)
		c.Status(204)
	})

	// 地块相关
	land.POST("/contracts/:id/plots", authMiddleware(), requirePermission("land:create"), func(c *gin.Context) {
		contractID, _ := strconv.Atoi(c.Param("id"))
		var contract LandContract
		if DB.First(&contract, contractID).Error != nil {
			c.JSON(404, gin.H{"detail": "合同不存在"})
			return
		}

		var req LandPlotCreateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		plot := LandPlot{
			ContractID:       contractID,
			PlotName:         req.PlotName,
			PlotCode:         req.PlotCode,
			ContractArea1994: req.ContractArea1994,
			ConfirmedArea:    req.ConfirmedArea,
			BoundaryEast:     req.BoundaryEast,
			BoundarySouth:    req.BoundarySouth,
			BoundaryWest:     req.BoundaryWest,
			BoundaryNorth:    req.BoundaryNorth,
		}
		DB.Create(&plot)

		// 更新合同地块数
		var count int64
		DB.Model(&LandPlot{}).Where("contract_id = ?", contractID).Count(&count)
		contract.TotalPlots = int(count)
		DB.Save(&contract)

		c.JSON(201, plot)
	})

	land.PUT("/plots/:id", authMiddleware(), requirePermission("land:update"), func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var plot LandPlot
		if DB.First(&plot, id).Error != nil {
			c.JSON(404, gin.H{"detail": "地块不存在"})
			return
		}

		var req LandPlotUpdateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(422, gin.H{"detail": err.Error()})
			return
		}

		updates := make(map[string]interface{})
		if req.PlotName != nil {
			updates["plot_name"] = *req.PlotName
		}
		if req.PlotCode != nil {
			updates["plot_code"] = *req.PlotCode
		}
		if req.ContractArea1994 != nil {
			updates["contract_area_1994"] = *req.ContractArea1994
		}
		if req.ConfirmedArea != nil {
			updates["confirmed_area"] = *req.ConfirmedArea
		}
		if req.BoundaryEast != nil {
			updates["boundary_east"] = *req.BoundaryEast
		}
		if req.BoundarySouth != nil {
			updates["boundary_south"] = *req.BoundarySouth
		}
		if req.BoundaryWest != nil {
			updates["boundary_west"] = *req.BoundaryWest
		}
		if req.BoundaryNorth != nil {
			updates["boundary_north"] = *req.BoundaryNorth
		}

		DB.Model(&plot).Updates(updates)
		c.JSON(200, plot)
	})

	land.DELETE("/plots/:id", authMiddleware(), requirePermission("land:delete"), func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var plot LandPlot
		if DB.First(&plot, id).Error != nil {
			c.JSON(404, gin.H{"detail": "地块不存在"})
			return
		}

		contractID := plot.ContractID
		DB.Delete(&plot)

		// 更新合同地块数
		var contract LandContract
		if DB.First(&contract, contractID).Error == nil {
			var count int64
			DB.Model(&LandPlot{}).Where("contract_id = ?", contractID).Count(&count)
			contract.TotalPlots = int(count)
			DB.Save(&contract)
		}

		c.Status(204)
	})
}
