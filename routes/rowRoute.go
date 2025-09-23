package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func RowRoutes(router *gin.Engine) {
	rowGroup := router.Group("/row")
	{
		rowGroup.POST("/add-rows", middleware.RequireLogin, controllers.AddRows)
	}
}
