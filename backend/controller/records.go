package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shuangseqiu/backend/models"
	"shuangseqiu/config"
	"shuangseqiu/pkg/response"
)

func ListRecords(c *gin.Context) {
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
	query := config.DB.WithContext(c.Request.Context()).Model(&models.SSQRecord{})
	if err := query.Count(&total).Error; err != nil {
		response.JSON(c, http.StatusInternalServerError, "查询开奖历史失败", nil)
		return
	}
	items := make([]models.SSQRecord, 0, limit)
	if err := query.Order("draw_date DESC, issue DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		response.JSON(c, http.StatusInternalServerError, "查询开奖历史失败", nil)
		return
	}
	response.JSON(c, http.StatusOK, "查询成功", gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}
