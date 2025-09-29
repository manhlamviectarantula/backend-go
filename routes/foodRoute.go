package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func FoodRoutes(router *gin.Engine) {
	foodGroup := router.Group("/food")
	{
		foodGroup.GET("/get-foods-of-branch/:BranchID", middleware.RequireLogin, controllers.GetFoodsOfBranch)
		foodGroup.POST("/add-food-of-branch/:BranchID", middleware.RequireLogin, controllers.AddFoodOfBranch)
		foodGroup.PUT("/update-food-of-branch/:FoodID", middleware.RequireLogin, controllers.UpdateFood)
		foodGroup.PUT("/change-food-status/:FoodID", middleware.RequireLogin, controllers.ChangeFoodStatus)

		foodGroup.GET("/get-foods-to-order/:BranchID", controllers.GetFoodsToOrder)
		// foodGroup.DELETE("/delete-food-of-branch/:FoodID", controllers.DeleteFood)
	}
}
