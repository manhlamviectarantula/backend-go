package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func MovieRoutes(router *gin.Engine) {
	movieGroup := router.Group("/movie")
	{
		movieGroup.POST("/add-movie", middleware.RequireLogin, controllers.AddMovie)
		movieGroup.PUT("/update-movie/:MovieID", middleware.RequireLogin, controllers.UpdateMovie)
		movieGroup.GET("/get-all-movie", middleware.RequireLogin, controllers.GetAllMovies)
		movieGroup.GET("/get-movies-in-add-showtime", middleware.RequireLogin, controllers.GetMoviesInAddShowtime)

		movieGroup.GET("/get-showing-movie", controllers.GetShowingMovie)
		movieGroup.GET("/get-upcoming-movie", controllers.GetUpcomingMovie)
		movieGroup.GET("/details-movie/:id", controllers.GetDetailsMovie)
	}
}
