package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func MessageRoutes(router *gin.Engine) {
	messageGroup := router.Group("/message")
	{
		messageGroup.GET("/get-users-sidebar-admin", middleware.RequireLogin, controllers.GetUsersSidebarAdmin)
		messageGroup.POST("/send-message/:userToChatID", middleware.RequireLogin, controllers.SendMessage)
		messageGroup.GET("/get-message/:userToChatID", middleware.RequireLogin, controllers.GetMessages)
	}
}
