package router

import (
	"github.com/gin-gonic/gin"
	"shuangseqiu/backend/controller"
	"shuangseqiu/middleware"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), gin.Recovery())
	r.GET("/health", controller.Health)
	api := r.Group("/api/v1")
	api.POST("/analysis", controller.Analyze)
	api.GET("/records", controller.ListRecords)
	api.GET("/fucai3d/records", controller.ListFucai3DRecords)
	api.POST("/fucai3d/analysis", controller.AnalyzeFucai3D)
	return r
}
