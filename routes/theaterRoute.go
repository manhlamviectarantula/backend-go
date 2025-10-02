package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func TheaterRoutes(router *gin.Engine) {
	theaterGroup := router.Group("/theater")
	{
		theaterGroup.GET("/get-all-theater-of-branch/:BranchID", middleware.RequireLogin, controllers.GetAllTheaterOfBranch)
		theaterGroup.GET("/get-details-theater/:TheaterID", middleware.RequireLogin, controllers.GetDetailsTheater)
		theaterGroup.GET("/get-seats-of-theater/:TheaterID", middleware.RequireLogin, controllers.GetSeatsOfTheater)
		theaterGroup.POST("/add-theater", middleware.RequireLogin, controllers.AddTheater)
		theaterGroup.PUT("/change-theater-status/:TheaterID", middleware.RequireLogin, controllers.ChangeTheaterStatus)
	}
}
