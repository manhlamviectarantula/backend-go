package routes

import (
	"movie-ticket-booking/controllers"

	"github.com/gin-gonic/gin"
)

func OrderFoodRoutes(router *gin.Engine) {
	OrderFoodGroup := router.Group("/orderFood")
	{
		OrderFoodGroup.POST("/add-order-food/:OrderID", controllers.AddOrderFood)
	}
}
