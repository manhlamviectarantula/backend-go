package controllers

import (
	"fmt"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonthlyTotal struct {
	Month string `json:"Month"`
	Total int    `json:"Total"`
}

func GetAllOrdersTotalAndCreatedAt(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tham số year"})
		return
	}

	// Truy vấn tổng theo tháng
	var results []struct {
		Month int
		Total int
	}

	query := `
		SELECT 
			MONTH(CreatedAt) AS Month,
			SUM(Total) AS Total
		FROM orders
		WHERE YEAR(CreatedAt) = ?
		GROUP BY MONTH(CreatedAt)
	`

	if err := database.DB.Raw(query, year).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn dữ liệu"})
		return
	}

	// Khởi tạo array có thứ tự tháng 1 đến 12
	monthlyTotals := make([]MonthlyTotal, 12)
	for i := 1; i <= 12; i++ {
		monthlyTotals[i-1] = MonthlyTotal{
			Month: fmt.Sprintf("Tháng %d", i),
			Total: 0,
		}
	}

	// Gán dữ liệu từ query vào slice
	for _, r := range results {
		index := r.Month - 1
		if index >= 0 && index < 12 {
			monthlyTotals[index].Total = r.Total
		}
	}

	c.JSON(http.StatusOK, monthlyTotals)
}

func GetPieChartAgeTag(c *gin.Context) {
	var results []struct {
		AgeTag     string  `json:"AgeTag"`
		Percentage float64 `json:"Percentage"`
	}

	query := `
		SELECT AgeTag, 
		       COUNT(*) * 100.0 / (SELECT COUNT(*) FROM movies) AS Percentage
		FROM movies
		GROUP BY AgeTag;
	`

	if err := database.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func GetMovieDropdown(c *gin.Context) {
	var movies []struct {
		MovieID   uint   `json:"MovieID"`
		MovieName string `json:"MovieName"`
	}

	// Chỉ lấy MovieID và MovieName
	if err := database.DB.Model(&models.Movie{}).
		Select("MovieID, MovieName").
		Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func GetMovieChart(c *gin.Context) {
	type Revenue struct {
		OrderDate        string  `json:"OrderDate"`
		TotalTicketPrice float64 `json:"TotalTicketPrice"`
	}

	MovieID := c.Query("MovieID")
	FromDate := c.Query("FromDate")
	ToDate := c.Query("ToDate")

	if MovieID == "" || FromDate == "" || ToDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "MovieID, FromDate, ToDate are required",
		})
		return
	}

	var revenues []Revenue
	query := `
        SELECT 
            DATE(o.CreatedAt) AS OrderDate,
            SUM(ss.TicketPrice) AS TotalTicketPrice
        FROM showtime_seats ss
        JOIN orders o ON o.OrderID = ss.OrderID
		JOIN showtimes s ON s.ShowtimeID = o.ShowtimeID 
		JOIN movies m ON m.MovieID = s.MovieID
        WHERE m.MovieID = ?
          AND DATE(o.CreatedAt) BETWEEN ? AND ?
        GROUP BY DATE(o.CreatedAt)
        ORDER BY OrderDate;
    `

	if err := database.DB.Raw(query, MovieID, FromDate, ToDate).Scan(&revenues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": revenues,
	})
}

func GetMovieOverall(c *gin.Context) {
	type Overview struct {
		SoldSeats      int     `json:"SoldSeats"`
		TotalRevenue   float64 `json:"TotalRevenue"`
		MinDay         string  `json:"MinDay"`
		MinDaySeats    int     `json:"MinDaySeats"`
		MinDayRevenue  float64 `json:"MinDayRevenue"`
		MaxDay         string  `json:"MaxDay"`
		MaxDaySeats    int     `json:"MaxDaySeats"`
		MaxDayRevenue  float64 `json:"MaxDayRevenue"`
		TotalShowtimes int     `json:"TotalShowtimes"`
	}

	MovieID := c.Query("MovieID")
	FromDate := c.Query("FromDate")
	ToDate := c.Query("ToDate")

	if MovieID == "" || FromDate == "" || ToDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "MovieID, FromDate, ToDate are required",
		})
		return
	}

	query := `
        SELECT 
            sold.sold_seats AS SoldSeats,
            rev.total_revenue AS TotalRevenue,
            minday.OrderDate AS MinDay,
            minday.TotalSeats AS MinDaySeats,
            minday.TotalTicketPrice AS MinDayRevenue,
            maxday.OrderDate AS MaxDay,
            maxday.TotalSeats AS MaxDaySeats,
            maxday.TotalTicketPrice AS MaxDayRevenue,
            stcount.TotalShowtimes AS TotalShowtimes
        FROM 
            (SELECT COUNT(*) AS sold_seats
             FROM showtime_seats ss
             JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
             JOIN orders o ON ss.OrderID = o.OrderID
             WHERE st.MovieID = ? AND ss.Status = 1
               AND DATE(o.CreatedAt) BETWEEN ? AND ?) sold
        CROSS JOIN
            (SELECT SUM(ss.TicketPrice) AS total_revenue
             FROM showtime_seats ss
             JOIN orders o ON ss.OrderID = o.OrderID
             JOIN showtimes st ON o.ShowtimeID = st.ShowtimeID
             WHERE st.MovieID = ? 
               AND DATE(o.CreatedAt) BETWEEN ? AND ?) rev
        CROSS JOIN
            (SELECT DATE(o.CreatedAt) AS OrderDate,
                    COUNT(ss.ShowtimeSeatID) AS TotalSeats,
                    SUM(ss.TicketPrice) AS TotalTicketPrice
             FROM showtime_seats ss
             JOIN orders o ON ss.OrderID = o.OrderID
             JOIN showtimes st ON o.ShowtimeID = st.ShowtimeID
             WHERE st.MovieID = ? 
               AND DATE(o.CreatedAt) BETWEEN ? AND ?
             GROUP BY DATE(o.CreatedAt)
             ORDER BY TotalTicketPrice ASC
             LIMIT 1) minday
        CROSS JOIN
            (SELECT DATE(o.CreatedAt) AS OrderDate,
                    COUNT(ss.ShowtimeSeatID) AS TotalSeats,
                    SUM(ss.TicketPrice) AS TotalTicketPrice
             FROM showtime_seats ss
             JOIN orders o ON ss.OrderID = o.OrderID
             JOIN showtimes st ON o.ShowtimeID = st.ShowtimeID
             WHERE st.MovieID = ? 
               AND DATE(o.CreatedAt) BETWEEN ? AND ?
             GROUP BY DATE(o.CreatedAt)
             ORDER BY TotalTicketPrice DESC
             LIMIT 1) maxday
        CROSS JOIN
            (SELECT COUNT(DISTINCT st.ShowtimeID) AS TotalShowtimes
             FROM showtimes st
             JOIN orders o ON st.ShowtimeID = o.ShowtimeID
             WHERE st.MovieID = ? 
               AND DATE(o.CreatedAt) BETWEEN ? AND ?) stcount;
    `

	var result Overview
	if err := database.DB.Raw(query,
		MovieID, FromDate, ToDate, // sold
		MovieID, FromDate, ToDate, // rev
		MovieID, FromDate, ToDate, // minday
		MovieID, FromDate, ToDate, // maxday
		MovieID, FromDate, ToDate, // stcount
	).Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func GetBranchDropdown(c *gin.Context) {
	var branches []struct {
		BranchID   uint   `json:"BranchID"`
		BranchName string `json:"BranchName"`
	}

	// Chỉ lấy BranchID và BranchName
	if err := database.DB.Model(&models.Branch{}).
		Select("BranchID, BranchName").
		Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

func GetBranchChart(c *gin.Context) {
	type Revenue struct {
		OrderDate        string  `json:"OrderDate"`
		TotalTicketPrice float64 `json:"TotalTicketPrice"`
	}

	BranchID := c.Query("BranchID")
	FromDate := c.Query("FromDate")
	ToDate := c.Query("ToDate")

	if BranchID == "" || FromDate == "" || ToDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "BranchID, FromDate, ToDate are required",
		})
		return
	}

	var revenues []Revenue
	query := `
	SELECT 
	DATE(o.CreatedAt) AS OrderDate,
    SUM(ss.TicketPrice) AS TotalTicketPrice
	FROM showtime_seats ss
	JOIN orders o ON ss.OrderID = o.OrderID
	JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
	JOIN theaters t ON st.TheaterID = t.TheaterID
	JOIN branches b ON t.BranchID = b.BranchID
	WHERE b.BranchID = ?
	  AND DATE(o.CreatedAt) BETWEEN ? AND ?
	GROUP BY DATE(o.CreatedAt)
	ORDER BY OrderDate;
    `

	if err := database.DB.Raw(query, BranchID, FromDate, ToDate).Scan(&revenues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": revenues,
	})
}

func GetBranchOverall(c *gin.Context) {
	type Overview struct {
		TotalSeats     int     `json:"TotalSeats"`
		TotalRevenue   float64 `json:"TotalRevenue"`
		MinSaleDate    string  `json:"MinSaleDate"`
		MinSaleSeats   int     `json:"MinSaleSeats"`
		MinSaleRevenue float64 `json:"MinSaleRevenue"`
		MaxSaleDate    string  `json:"MaxSaleDate"`
		MaxSaleSeats   int     `json:"MaxSaleSeats"`
		MaxSaleRevenue float64 `json:"MaxSaleRevenue"`
		TotalShowtimes int     `json:"TotalShowtimes"`
	}

	branchID := c.Query("BranchID")
	fromDate := c.Query("FromDate")
	toDate := c.Query("ToDate")

	if branchID == "" || fromDate == "" || toDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BranchID, FromDate, ToDate are required"})
		return
	}

	query := `
		SELECT 
			sold.sold_seats AS TotalSeats,
			rev.total_revenue AS TotalRevenue,
			minday.OrderDate AS MinSaleDate,
			minday.TotalSeats AS MinSaleSeats,
			minday.TotalTicketPrice AS MinSaleRevenue,
			maxday.OrderDate AS MaxSaleDate,
			maxday.TotalSeats AS MaxSaleSeats,
			maxday.TotalTicketPrice AS MaxSaleRevenue,
			stcount.TotalShowtimes AS TotalShowtimes
		FROM 
			(SELECT COUNT(ss.ShowtimeSeatID) AS sold_seats
			 FROM showtime_seats ss
			 JOIN orders o ON ss.OrderID = o.OrderID
			 JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
			 JOIN theaters t ON st.TheaterID = t.TheaterID
			 JOIN branches b ON t.BranchID = b.BranchID
			 WHERE b.BranchID = ? AND DATE(o.CreatedAt) BETWEEN ? AND ?) sold
		CROSS JOIN
			(SELECT IFNULL(SUM(ss.TicketPrice),0) AS total_revenue
			 FROM showtime_seats ss
			 JOIN orders o ON ss.OrderID = o.OrderID
			 JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
			 JOIN theaters t ON st.TheaterID = t.TheaterID
			 JOIN branches b ON t.BranchID = b.BranchID
			 WHERE b.BranchID = ? AND DATE(o.CreatedAt) BETWEEN ? AND ?) rev
		CROSS JOIN
			(SELECT DATE(o.CreatedAt) AS OrderDate,
					COUNT(ss.ShowtimeSeatID) AS TotalSeats,
					SUM(ss.TicketPrice) AS TotalTicketPrice
			 FROM showtime_seats ss
			 JOIN orders o ON ss.OrderID = o.OrderID
			 JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
			 JOIN theaters t ON st.TheaterID = t.TheaterID
			 JOIN branches b ON t.BranchID = b.BranchID
			 WHERE b.BranchID = ? AND DATE(o.CreatedAt) BETWEEN ? AND ?
			 GROUP BY DATE(o.CreatedAt)
			 ORDER BY TotalTicketPrice ASC
			 LIMIT 1) minday
		CROSS JOIN
			(SELECT DATE(o.CreatedAt) AS OrderDate,
					COUNT(ss.ShowtimeSeatID) AS TotalSeats,
					SUM(ss.TicketPrice) AS TotalTicketPrice
			 FROM showtime_seats ss
			 JOIN orders o ON ss.OrderID = o.OrderID
			 JOIN showtimes st ON ss.ShowtimeID = st.ShowtimeID
			 JOIN theaters t ON st.TheaterID = t.TheaterID
			 JOIN branches b ON t.BranchID = b.BranchID
			 WHERE b.BranchID = ? AND DATE(o.CreatedAt) BETWEEN ? AND ?
			 GROUP BY DATE(o.CreatedAt)
			 ORDER BY TotalTicketPrice DESC
			 LIMIT 1) maxday
		CROSS JOIN
			(SELECT COUNT(DISTINCT st.ShowtimeID) AS TotalShowtimes
			 FROM showtimes st
			 JOIN theaters t ON st.TheaterID = t.TheaterID
			 JOIN branches b ON t.BranchID = b.BranchID
			 JOIN orders o ON st.ShowtimeID = o.ShowtimeID
			 WHERE b.BranchID = ? AND DATE(o.CreatedAt) BETWEEN ? AND ?) stcount;
	`

	var result Overview
	if err := database.DB.Raw(query,
		branchID, fromDate, toDate, // sold
		branchID, fromDate, toDate, // rev
		branchID, fromDate, toDate, // minday
		branchID, fromDate, toDate, // maxday
		branchID, fromDate, toDate, // stcount
	).Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// BranchAdmin
func GetFoodDropdown(c *gin.Context) {
	// Lấy BranchID từ URL
	branchID := c.Param("BranchID")

	var foods []struct {
		FoodID   uint   `json:"FoodID"`
		FoodName string `json:"FoodName"`
	}

	// Lọc theo BranchID
	if err := database.DB.Model(&models.Food{}).
		Select("FoodID, FoodName").
		Where("BranchID = ?", branchID).
		Find(&foods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch foods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"foods": foods})
}

func GetFoodChart(c *gin.Context) {
	type FoodRevenue struct {
		OrderDate  string  `json:"OrderDate"`
		TotalPrice float64 `json:"TotalPrice"`
	}

	FoodID := c.Query("FoodID")
	FromDate := c.Query("FromDate")
	ToDate := c.Query("ToDate")

	if FoodID == "" || FromDate == "" || ToDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "FoodID, FromDate, ToDate are required",
		})
		return
	}

	var revenues []FoodRevenue
	query := `
	SELECT 
		DATE(o.CreatedAt) AS OrderDate,
		SUM(ofs.TotalPrice) AS TotalPrice
	FROM order_foods AS ofs
	JOIN orders AS o ON ofs.OrderID = o.OrderID
	WHERE ofs.FoodID = ?
	  AND DATE(o.CreatedAt) BETWEEN ? AND ?
	GROUP BY DATE(o.CreatedAt)
	ORDER BY OrderDate;
	`

	if err := database.DB.Raw(query, FoodID, FromDate, ToDate).Scan(&revenues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": revenues,
	})
}

func GetFoodOverall(c *gin.Context) {
	type FoodOverview struct {
		BestSellingName    string  `json:"BestSellingName"`
		BestSellingQty     int     `json:"BestSellingQty"`
		LeastSellingName   string  `json:"LeastSellingName"`
		LeastSellingQty    int     `json:"LeastSellingQty"`
		MostExpensiveName  string  `json:"MostExpensiveName"`
		MostExpensivePrice float64 `json:"MostExpensivePrice"`
		CheapestName       string  `json:"CheapestName"`
		CheapestPrice      float64 `json:"CheapestPrice"`
		TotalRevenue       float64 `json:"TotalRevenue"`
	}

	branchID := c.Param("BranchID")
	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BranchID is required"})
		return
	}

	var overall FoodOverview

	// 1. Món bán chạy nhất
	type BestSelling struct {
		Name string
		Qty  int
	}
	var bs BestSelling
	database.DB.Raw(`
		SELECT f.FoodName AS Name, SUM(ofs.Quantity) AS Qty
		FROM order_foods ofs
		JOIN foods f ON f.FoodID = ofs.FoodID
		WHERE f.BranchID = ?
		GROUP BY f.FoodID, f.FoodName
		ORDER BY Qty DESC
		LIMIT 1
	`, branchID).Scan(&bs)
	overall.BestSellingName = bs.Name
	overall.BestSellingQty = bs.Qty

	// 2. Món bán ít nhất
	type LeastSelling struct {
		Name string
		Qty  int
	}
	var ls LeastSelling
	database.DB.Raw(`
		SELECT f.FoodName AS Name, SUM(ofs.Quantity) AS Qty
		FROM order_foods ofs
		JOIN foods f ON f.FoodID = ofs.FoodID
		WHERE f.BranchID = ?
		GROUP BY f.FoodID, f.FoodName
		ORDER BY Qty ASC
		LIMIT 1
	`, branchID).Scan(&ls)
	overall.LeastSellingName = ls.Name
	overall.LeastSellingQty = ls.Qty

	// 3. Món giá cao nhất
	type Expensive struct {
		Name  string
		Price float64
	}
	var exp Expensive
	database.DB.Raw(`
		SELECT FoodName AS Name, Price
		FROM foods
		WHERE BranchID = ?
		ORDER BY Price DESC
		LIMIT 1
	`, branchID).Scan(&exp)
	overall.MostExpensiveName = exp.Name
	overall.MostExpensivePrice = exp.Price

	// 4. Món rẻ nhất
	var cheap Expensive
	database.DB.Raw(`
		SELECT FoodName AS Name, Price
		FROM foods
		WHERE BranchID = ?
		ORDER BY Price ASC
		LIMIT 1
	`, branchID).Scan(&cheap)
	overall.CheapestName = cheap.Name
	overall.CheapestPrice = cheap.Price

	// 5. Tổng doanh thu
	type Total struct {
		Revenue float64
	}
	var tot Total
	database.DB.Raw(`
		SELECT SUM(ofs.TotalPrice) AS Revenue
		FROM foods f
		JOIN order_foods ofs ON f.FoodID = ofs.FoodID
		WHERE f.BranchID = ?
	`, branchID).Scan(&tot)
	overall.TotalRevenue = tot.Revenue

	c.JSON(http.StatusOK, gin.H{"data": overall})
}
