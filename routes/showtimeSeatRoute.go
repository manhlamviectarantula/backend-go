package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func ShowtimeSeatRoutes(router *gin.Engine) {
	showtimeSeatGroup := router.Group("/showtime-seat")
	{
		showtimeSeatGroup.POST("/add-showtime-seats/:ShowtimeID/:TheaterID", middleware.RequireLogin, controllers.AddShowtimeSeats)
		showtimeSeatGroup.DELETE("/delete-showtime-seats/:ShowtimeID", middleware.RequireLogin, controllers.DeleteShowtimeSeats)

		showtimeSeatGroup.GET("/get-seat-of-showtime", controllers.GetSeatOfShowtime)
	}
}
