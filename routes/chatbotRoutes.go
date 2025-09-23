package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func ChatbotRoutes(router *gin.Engine) {
	chatbotGroup := router.Group("/chatbot")
	{
		chatbotGroup.POST("/ai", middleware.RequireLogin, controllers.ChatAIHandler)
		chatbotGroup.POST("/classic", middleware.RequireLogin, controllers.ChatClassicHandler)
	}
}
