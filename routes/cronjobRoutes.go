package routes

import (
	"movie-ticket-booking/services"

	"github.com/gin-gonic/gin"
)

func CronjobRoutes(router *gin.Engine) {
	cronjobGroup := router.Group("/cronjob")
	{
		// cronjobGroup.PUT("/update-movie-status", middleware.RequireLogin, controllers.GetFoodsOfBranch)
		cronjobGroup.PUT("/update-movie-status", services.DailyUpdateMoviesHandler)
		cronjobGroup.PUT("/unlock-seat", services.AutoUnlockSeatsHandler)
		cronjobGroup.PUT("/close-showtime", services.AutoCloseShowtimesHandler)
	}
}
