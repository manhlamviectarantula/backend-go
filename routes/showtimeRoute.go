package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func ShowtimeRoutes(router *gin.Engine) {
	showtimeGroup := router.Group("/showtime")
	{
		showtimeGroup.GET("/get-all-showtimes-of-branch/:BranchID", middleware.RequireLogin, controllers.GetAllShowtimesOfBranch)
		showtimeGroup.POST("/add-showtime", middleware.RequireLogin, controllers.AddShowtime)
		showtimeGroup.GET("/get-details-showtime/:ShowtimeID", middleware.RequireLogin, controllers.GetDetailsShowtime)
		showtimeGroup.PUT("/open-order-showtime/:ShowtimeID", middleware.RequireLogin, controllers.OpenOrderShowtime)
		showtimeGroup.PUT("/cancel-showtime/:ShowtimeID", middleware.RequireLogin, controllers.CancelShowtime)
		showtimeGroup.DELETE("/delete-showtime/:ShowtimeID", middleware.RequireLogin, controllers.DeleteShowtime)

		showtimeGroup.GET("/get-showtimes-of-date/:MovieID", controllers.GetAllShowtimesOfDate)
		showtimeGroup.GET("/get-showtimes-info-in-selectSeat/:ShowtimeID", controllers.GetShowtimeInfo)

		showtimeGroup.PUT("/test-cron", controllers.TestCron)

	}
}
