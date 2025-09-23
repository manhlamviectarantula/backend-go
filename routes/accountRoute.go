package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func AccountRoutes(router *gin.Engine) {
	accountGroup := router.Group("/account")
	{
		accountGroup.GET("/get-all-accounts", middleware.RequireLogin, controllers.GetAllAccounts)
		accountGroup.PUT("/change-account-status/:AccountID", middleware.RequireLogin, controllers.BlockAccount)
		accountGroup.PUT("/upgrade-account", middleware.RequireLogin, controllers.UpgradeAccount)
		accountGroup.PUT("/downgrade-account", middleware.RequireLogin, controllers.DowngradeAccount)
		accountGroup.PUT("/update-account/:AccountID", middleware.RequireLogin, controllers.UpdateAccount)
		accountGroup.PUT("/update-pw/:AccountID", middleware.RequireLogin, controllers.UpdatePassword)

		accountGroup.POST("/forget-pw", controllers.ForgetPassword)
	}
}
