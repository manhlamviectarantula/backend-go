package main

import (
	"log"
	"movie-ticket-booking/config"
	"movie-ticket-booking/database"
	"movie-ticket-booking/routes"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	database.Connect()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "cron-update-movies":
			UpdateMovies()
		case "cron-unlock-seats":
			AutoUnlockSeats()
		case "cron-daily-movies":
			DailyUpdateMovies()
		case "cron-close-showtimes":
			AutoCloseShowtimes()
		case "cron-all": // chạy tất cả cron liên tục
			c := SetupCronJobs()
			c.Start()
			defer c.Stop()
			select {} // giữ process chạy
		}
		return
	}

	// Tạo socket server
	socketServer := InitSocket()
	go socketServer.Serve()
	defer socketServer.Close()

	c := SetupCronJobs()
	c.Start()
	defer c.Stop()

	router := gin.Default()

	// CORS middleware
	origins := strings.Split(config.GetEnv("CORS_ALLOW_ORIGINS", ""), ",")
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Tích hợp socket.io
	router.GET("/socket.io/*any", gin.WrapH(socketServer))
	router.POST("/socket.io/*any", gin.WrapH(socketServer))

	// Đăng ký các routes
	routes.MovieRoutes(router)
	routes.AuthRoutes(router)
	routes.BranchRoutes(router)
	routes.TheaterRoutes(router)
	routes.RowRoutes(router)
	routes.SeatRoutes(router)
	routes.ShowtimeRoutes(router)
	routes.AccountRoutes(router)
	routes.ShowtimeSeatRoutes(router)
	routes.OrderRoutes(router)
	routes.FoodRoutes(router)
	routes.OrderFoodRoutes(router)
	routes.MessageRoutes(router)
	routes.AdminDashboardRoutes(router)
	routes.ChatbotRoutes(router)

	port := config.GetEnv("PORT", "8080")
	log.Println("✅ Server đang chạy tại cổng " + port + "...")
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}
