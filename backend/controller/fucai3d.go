package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"shuangseqiu/backend/models"
	"shuangseqiu/backend/service"
	"shuangseqiu/config"
	"shuangseqiu/pkg/response"
	"strconv"
)

func AnalyzeFucai3D(c *gin.Context) {
	var req struct {
		Numbers []int `json:"numbers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, "参数错误：需要百、十、个位共 3 个数字", nil)
		return
	}
	result, err := service.AnalyzeFucai3D(req.Numbers)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.JSON(c, http.StatusOK, "分析成功", result)
}

func ListFucai3DRecords(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if err != nil || limit < 1 || limit > 500 {
		response.JSON(c, http.StatusBadRequest, "limit 必须在 1–500 之间", nil)
		return
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 100000 {
		response.JSON(c, http.StatusBadRequest, "offset 必须在 0–100000 之间", nil)
		return
	}
	if config.DB == nil {
		response.JSON(c, http.StatusServiceUnavailable, "数据库未启用，无法查询开奖历史", nil)
		return
	}
	var total int64
	query := config.DB.WithContext(c.Request.Context()).Model(&models.Fucai3DRecord{})
	if err := query.Count(&total).Error; err != nil {
		response.JSON(c, http.StatusInternalServerError, "查询福彩3D开奖历史失败", nil)
		return
	}
	items := make([]models.Fucai3DRecord, 0, limit)
	if err := query.Order("draw_date DESC, issue DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		response.JSON(c, http.StatusInternalServerError, "查询福彩3D开奖历史失败", nil)
		return
	}
	response.JSON(c, http.StatusOK, "查询成功", gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}
