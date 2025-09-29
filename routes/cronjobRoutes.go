package routes

import (
	"movie-ticket-booking/services"

	"github.com/gin-gonic/gin"
)

func CronjobRoutes(router *gin.Engine) {
	cronjobGroup := router.Group("/cronjob")
	{
		// cronjobGroup.POST("/update-movie-status", middleware.RequireLogin, controllers.GetFoodsOfBranch)
		cronjobGroup.POST("/update-movie-status", services.DailyUpdateMoviesHandler)
		cronjobGroup.POST("/unlock-seat", services.AutoUnlockSeatsHandler)
		cronjobGroup.POST("/close-showtime", services.AutoCloseShowtimesHandler)
	}
}
