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
		branchGroup.GET("/get-details-branch/:BranchID", controllers.GetDetailsBranch)
		branchGroup.POST("/add-branch", middleware.RequireLogin, controllers.AddBranch)
		branchGroup.PUT("/update-branch/:BranchID", controllers.UpdateBranch)
	}
}
