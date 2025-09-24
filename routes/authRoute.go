package routes

import (
	"movie-ticket-booking/controllers"
	"movie-ticket-booking/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", controllers.RegisterUser)
		authGroup.POST("/login", controllers.LoginUser)
		authGroup.GET("/facebook/login", controllers.FacebookLogin)
		authGroup.GET("/facebook/callback", controllers.FacebookCallback)
		authGroup.GET("/google/login", controllers.GoogleLogin)
		authGroup.GET("/google/callback", controllers.GoogleCallback)

		authGroup.GET("/check", middleware.RequireLogin, controllers.CheckAuth)
		authGroup.GET("/tMidd", middleware.RequireLogin, controllers.TMidd)

		authGroup.POST("/request-otp/:email", controllers.RequestOTP)
		authGroup.POST("/verify-otp", controllers.VerifyOTP)
	}
}
