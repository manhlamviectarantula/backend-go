package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.Engine) {
	orderGroup := router.Group("/order")
	{
		orderGroup.GET("/get-orders-of-account/:AccountID", middleware.RequireLogin, controllers.GetOrdersOfAccount)

		orderGroup.POST("/add-order", controllers.AddOrder)
		orderGroup.POST("/create-payment", controllers.CreateMomoPayment)
		// orderGroup.POST("/result-payment", controllers.MomoResultHandler)
		orderGroup.POST("/create-after-payment", controllers.CreateOrderAfterPayment)

	}
}
