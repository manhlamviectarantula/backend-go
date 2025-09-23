package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func SeatRoutes(router *gin.Engine) {
	seatGroup := router.Group("/seat")
	{
		seatGroup.POST("/add-seats", middleware.RequireLogin, controllers.AddSeats)
	}
}
