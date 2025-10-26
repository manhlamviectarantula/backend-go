package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func BranchRoutes(router *gin.Engine) {
	branchGroup := router.Group("/branch")
	{
		branchGroup.GET("/get-all-branch", controllers.GetAllBranch)
		branchGroup.POST("/add-branch", middleware.RequireLogin, controllers.AddBranch)

		branchGroup.GET("/get-details-branch/:BranchID", middleware.RequireLogin, controllers.GetDetailsBranch)
		branchGroup.PUT("/update-branch/:BranchID", middleware.RequireLogin, controllers.UpdateBranch)
	}
}
