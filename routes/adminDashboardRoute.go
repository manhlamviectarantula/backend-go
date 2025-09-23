package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func AdminDashboardRoutes(router *gin.Engine) {
	AdminDashboardGroup := router.Group("/adminDashboard")
	{
		AdminDashboardGroup.GET("/get-order-total-chart", middleware.RequireLogin, controllers.GetAllOrdersTotalAndCreatedAt)
		AdminDashboardGroup.GET("/get-pie-chart", middleware.RequireLogin, controllers.GetPieChartAgeTag)
		AdminDashboardGroup.GET("/get-movie-dropdown", middleware.RequireLogin, controllers.GetMovieDropdown)
		AdminDashboardGroup.GET("/get-movie-chart", middleware.RequireLogin, controllers.GetMovieChart)
		AdminDashboardGroup.GET("/get-movie-overall", middleware.RequireLogin, controllers.GetMovieOverall)
		AdminDashboardGroup.GET("/get-branch-dropdown", middleware.RequireLogin, controllers.GetBranchDropdown)
		AdminDashboardGroup.GET("/get-branch-chart", middleware.RequireLogin, controllers.GetBranchChart)
		AdminDashboardGroup.GET("/get-branch-overall", middleware.RequireLogin, controllers.GetBranchOverall)

		AdminDashboardGroup.GET("get-food-dropdown/:BranchID", middleware.RequireLogin, controllers.GetFoodDropdown)
		AdminDashboardGroup.GET("/get-food-chart", middleware.RequireLogin, controllers.GetFoodChart)
		AdminDashboardGroup.GET("/get-food-overall/:BranchID", middleware.RequireLogin, controllers.GetFoodOverall)
	}
}
