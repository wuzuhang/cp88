package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shuangseqiu/backend/controller/dto"
	"shuangseqiu/backend/service"
	"shuangseqiu/pkg/response"
)

var analysisService = &service.AnalysisService{}

func Analyze(c *gin.Context) {
	var req dto.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, "参数错误：需要 6 个红球和 1 个蓝球", nil)
		return
	}
	result, err := analysisService.Analyze(req)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.JSON(c, http.StatusOK, "分析成功", result)
}

func Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, "ok", gin.H{"service": "shuangseqiu"})
}
