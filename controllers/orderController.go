package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"movie-ticket-booking/config"
	"movie-ticket-booking/database"
	"movie-ticket-booking/models"
	"movie-ticket-booking/services"
	"movie-ticket-booking/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

func GetOrdersOfAccount(c *gin.Context) {
	accountID := c.Param("AccountID")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AccountID is required"})
		return
	}

	type SeatInfo struct {
		RowName     string  `json:"RowName"`
		SeatNumber  string  `json:"SeatNumber"`
		TicketPrice float64 `json:"TicketPrice"`
	}

	type FoodInfo struct {
		FoodName    string  `json:"FoodName"`
		Description string  `json:"Description"`
		Price       float64 `json:"Price"`
		Quantity    int     `json:"Quantity"`
	}

	type TicketRow struct {
		OrderID     int
		MovieName   string
		TheaterName string
		BranchName  string
		StartTime   string
		ShowDate    string
		TicketPrice float64
		RowName     string
		SeatNumber  string
		Total       float64
		CreatedAt   time.Time
	}

	// Query lấy toàn bộ ghế theo AccountID
	var ticketRows []TicketRow
	if err := database.DB.Raw(`
	SELECT 
		o.OrderID,
		m.MovieName,
		t.TheaterName,
		b.BranchName,
		s.StartTime,
		s.ShowDate,
		ss.TicketPrice,
		ss.RowName,
		se.SeatNumber,
		o.Total,
		o.CreatedAt
	FROM orders o
	JOIN accounts a ON a.AccountID = o.AccountID
	JOIN showtimes s ON s.ShowtimeID = o.ShowtimeID
	JOIN movies m ON m.MovieID = s.MovieID
	JOIN showtime_seats ss ON ss.OrderID = o.OrderID
	JOIN seats se ON se.SeatID = ss.SeatID
	JOIN theaters t ON t.TheaterID = s.TheaterID
	JOIN branches b ON b.BranchID = t.BranchID
	WHERE a.AccountID = ?
	ORDER BY o.OrderID
`, accountID).Scan(&ticketRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ticket info"})
		return
	}

	// Lấy danh sách OrderID
	orderIDs := make([]int, 0)
	for _, row := range ticketRows {
		if !contains(orderIDs, row.OrderID) {
			orderIDs = append(orderIDs, row.OrderID)
		}
	}

	// Query lấy foods
	foodsMap := make(map[int][]FoodInfo)
	if len(orderIDs) > 0 {
		var foods []struct {
			OrderID     int
			FoodName    string
			Description string
			Price       float64
			Quantity    int
		}
		if err := database.DB.Raw(`
			SELECT 
				o.OrderID,
				f.FoodName,
				f.Description,
				f.Price,
				ofs.Quantity
			FROM orders o
			JOIN order_foods ofs ON ofs.OrderID = o.OrderID
			JOIN foods f ON f.FoodID = ofs.FoodID
			WHERE o.OrderID IN ?
		`, orderIDs).Scan(&foods).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get foods"})
			return
		}

		for _, f := range foods {
			foodsMap[f.OrderID] = append(foodsMap[f.OrderID], FoodInfo{
				FoodName:    f.FoodName,
				Description: f.Description,
				Price:       f.Price,
				Quantity:    f.Quantity,
			})
		}
	}

	// Group dữ liệu theo OrderID
	type OrderResponse struct {
		OrderID     int        `json:"OrderID"`
		MovieName   string     `json:"MovieName"`
		TheaterName string     `json:"TheaterName"`
		BranchName  string     `json:"BranchName"`
		StartTime   string     `json:"StartTime"`
		ShowDate    string     `json:"ShowDate"`
		Total       float64    `json:"Total"`
		CreatedAt   time.Time  `json:"CreatedAt"`
		Foods       []FoodInfo `json:"Foods"`
		Seats       []SeatInfo `json:"Seats"`
	}

	orderMap := make(map[int]*OrderResponse)
	for _, row := range ticketRows {
		if _, exists := orderMap[row.OrderID]; !exists {
			orderMap[row.OrderID] = &OrderResponse{
				OrderID:     row.OrderID,
				MovieName:   row.MovieName,
				TheaterName: row.TheaterName,
				BranchName:  row.BranchName,
				StartTime:   row.StartTime,
				ShowDate:    row.ShowDate,
				Total:       row.Total,
				CreatedAt:   row.CreatedAt,
				Foods:       foodsMap[row.OrderID],
				Seats:       []SeatInfo{},
			}
		}
		orderMap[row.OrderID].Seats = append(orderMap[row.OrderID].Seats, SeatInfo{
			RowName:     row.RowName,
			SeatNumber:  row.SeatNumber,
			TicketPrice: row.TicketPrice,
		})
	}

	// Chuyển map -> slice
	result := make([]OrderResponse, 0, len(orderMap))
	for _, order := range orderMap {
		result = append(result, *order)
	}

	c.JSON(http.StatusOK, gin.H{"orders": result})
}

func contains(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func AddOrder(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.CreatedAt = time.Now()

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"orderID": order.OrderID,
	})
}

func CreateMomoPayment(c *gin.Context) {
	var request struct {
		Order              models.Order       `json:"order"`
		OrderFoods         []models.OrderFood `json:"orderFoods"`
		ShowtimeSeatUpdate struct {
			ShowtimeSeatIDs []int `json:"ShowtimeSeatIDs"`
		} `json:"showtimeSeatUpdates"`
	}

	// Parse request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// ✅ Kiểm tra suất chiếu còn hợp lệ không
	var showtime models.Showtime
	if err := database.DB.First(&showtime, request.Order.ShowtimeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy suất chiếu"})
		return
	}

	layout := "2006-01-02 15:04"
	showtimeStartStr := fmt.Sprintf("%s %s", showtime.ShowDate, showtime.StartTime)
	showtimeStartTime, err := time.ParseInLocation(layout, showtimeStartStr, time.Local)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Định dạng ngày/giờ suất chiếu không hợp lệ"})
		return
	}

	if !time.Now().Before(showtimeStartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Suất chiếu đã đóng đặt vé"})
		return
	}

	// -- Encode request data into extraData --
	rawData, _ := json.Marshal(request)
	extraData := base64.StdEncoding.EncodeToString(rawData)

	// -- MoMo config --
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	orderIDGen, _ := flake.NextID()
	requestIDGen, _ := flake.NextID()

	endpoint := "https://test-payment.momo.vn/v2/gateway/api/create"
	momoCfg := config.GetMomoEnv()
	partnerCode := momoCfg["PARTNER_CODE"]
	accessKey := momoCfg["ACCESS_KEY"]
	secretKey := momoCfg["SECRET_KEY"]
	redirectUrl := momoCfg["REDIRECT_URL"]
	ipnUrl := momoCfg["IPN_URL"]
	amount := strconv.Itoa(int(request.Order.Total))
	orderId := strconv.FormatUint(orderIDGen, 10)
	requestId := strconv.FormatUint(requestIDGen, 10)
	orderInfo := "Thanh toán vé xem phim tại CINÉMÀ"
	requestType := "payWithMethod"

	// -- Signature --
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=" + accessKey)
	rawSignature.WriteString("&amount=" + amount)
	rawSignature.WriteString("&extraData=" + extraData)
	rawSignature.WriteString("&ipnUrl=" + ipnUrl)
	rawSignature.WriteString("&orderId=" + orderId)
	rawSignature.WriteString("&orderInfo=" + orderInfo)
	rawSignature.WriteString("&partnerCode=" + partnerCode)
	rawSignature.WriteString("&redirectUrl=" + redirectUrl)
	rawSignature.WriteString("&requestId=" + requestId)
	rawSignature.WriteString("&requestType=" + requestType)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(rawSignature.Bytes())
	signature := hex.EncodeToString(h.Sum(nil))

	// -- Payload --
	payload := map[string]interface{}{
		"partnerCode":  partnerCode,
		"accessKey":    accessKey,
		"requestId":    requestId,
		"amount":       amount,
		"orderId":      orderId,
		"orderInfo":    orderInfo,
		"redirectUrl":  redirectUrl,
		"ipnUrl":       ipnUrl,
		"extraData":    extraData,
		"requestType":  requestType,
		"signature":    signature,
		"lang":         "vi",
		"autoCapture":  true,
		"orderGroupId": "",
		"partnerName":  "Movie Ticket",
		"storeId":      "MT001",
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MoMo payment"})
		return
	}
	defer resp.Body.Close()

	var momoResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&momoResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse MoMo response"})
		return
	}

	payUrl, ok := momoResp["payUrl"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid MoMo response", "response": momoResp})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payUrl": payUrl,
	})
}

func CreateOrderAfterPayment(c *gin.Context) {
	var request struct {
		Order              models.Order       `json:"order"`
		OrderFoods         []models.OrderFood `json:"orderFoods"`
		ShowtimeSeatUpdate struct {
			ShowtimeSeatIDs []int `json:"ShowtimeSeatIDs"`
		} `json:"showtimeSeatUpdates"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Lưu order
	request.Order.CreatedAt = time.Now()
	if err := database.DB.Create(&request.Order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order"})
		return
	}

	// Nếu có AccountID thì cộng thêm Point = Total
	if request.Order.AccountID != 0 {
		if err := database.DB.Model(&models.Account{}).
			Where("AccountID = ?", request.Order.AccountID).
			Update("Point", gorm.Expr("Point + ?", request.Order.Total)).Error; err != nil {
			log.Printf("❌ Cộng điểm thất bại cho Account %d: %v", request.Order.AccountID, err)
		}
	}

	// Lưu order foods
	for _, food := range request.OrderFoods {
		food.OrderID = request.Order.OrderID
		if err := database.DB.Create(&food).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order food"})
			return
		}
	}

	// Update seats
	for _, seatID := range request.ShowtimeSeatUpdate.ShowtimeSeatIDs {
		var seat models.ShowtimeSeat
		if err := database.DB.First(&seat, seatID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seat not found", "seatID": seatID})
			return
		}

		seat.Status = 2
		seat.OrderID = request.Order.OrderID

		if err := database.DB.Save(&seat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update seat"})
			return
		}
	}

	// Gửi mail nếu là khách vãng lai
	if request.Order.Email != "" {
		go func(orderID int) {
			if err := SendOrderInvoiceByID(orderID); err != nil {
				log.Printf("❌ Gửi email thất bại cho order %d: %v", orderID, err)
			}
		}(request.Order.OrderID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order saved successfully"})
}

func SendOrderInvoiceByID(orderID int) error {
	var order struct {
		OrderID     int
		Email       string
		MovieName   string
		TheaterName string
		BranchName  string
		ShowDate    string
		StartTime   string
		Total       int
	}
	if err := database.DB.
		Table("orders o").
		Select(`o.OrderID, o.Email, m.MovieName, t.TheaterName, 
            b.BranchName, s.ShowDate, s.StartTime, o.Total`).
		Joins("JOIN showtimes s ON s.ShowtimeID = o.ShowtimeID").
		Joins("JOIN movies m ON m.MovieID = s.MovieID").
		Joins("JOIN theaters t ON t.TheaterID = s.TheaterID").
		Joins("JOIN branches b ON b.BranchID = t.BranchID").
		Where("o.OrderID = ?", orderID).
		Scan(&order).Error; err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if order.Email == "" {
		return fmt.Errorf("no email provided")
	}

	// Lấy danh sách ghế
	var seats []struct {
		RowName     string
		SeatNumber  string
		TicketPrice int
	}
	if err := database.DB.Raw(`
		SELECT 
			ss.RowName,
			se.SeatNumber,
			ss.TicketPrice
		FROM orders o
		JOIN showtime_seats ss ON ss.OrderID = o.OrderID
		JOIN seats se ON se.SeatID = ss.SeatID
		WHERE o.OrderID = ?
	`, order.OrderID).Scan(&seats).Error; err != nil {
		return fmt.Errorf("failed to fetch seats: %v", err)
	}

	// Lấy danh sách món ăn
	var foods []struct {
		FoodName    string
		Description string
		Price       int
		Quantity    int
	}
	if err := database.DB.Raw(`
		SELECT 
			f.FoodName,
			f.Description,
			f.Price,
			ofs.Quantity
		FROM orders o
		JOIN order_foods ofs ON ofs.OrderID = o.OrderID
		JOIN foods f ON f.FoodID = ofs.FoodID
		WHERE o.OrderID = ?
	`, order.OrderID).Scan(&foods).Error; err != nil {
		return fmt.Errorf("failed to fetch order foods: %v", err)
	}

	// Tạo QR code
	ticketCode := utils.GenerateTicketCode(10)
	qrImage, err := utils.GenerateQRCode(ticketCode)
	if err != nil {
		return fmt.Errorf("failed to generate QR code")
	}

	// HTML ghế
	var seatHTML string
	for _, s := range seats {
		seatHTML += fmt.Sprintf("%s%s - %dđ<br/>", s.RowName, s.SeatNumber, s.TicketPrice)
	}

	// HTML món ăn
	var foodHTML string
	if len(foods) > 0 {
		foodHTML += "<h3>🍿 Thức ăn kèm theo:</h3><ul>"
		for _, f := range foods {
			foodHTML += fmt.Sprintf("<li>%s (%s) - %dđ x %d</li>",
				f.FoodName, f.Description, f.Price, f.Quantity)
		}
		foodHTML += "</ul>"
	}

	// Nội dung email
	subject := "🎟️ Hóa đơn đặt vé xem phim từ CINÉMÀ"
	body := fmt.Sprintf(`
		<h2>Cảm ơn bạn đã đặt vé!</h2>
		<p><strong>Phim:</strong> %s</p>
		<p><strong>Rạp:</strong> %s - %s</p>
		<p><strong>Ngày chiếu:</strong> %s</p>
		<p><strong>Giờ chiếu:</strong> %s</p>
		<p><strong>Ghế:</strong><br/> %s</p>
		%s
		<p><strong>Tổng cộng:</strong> %dđ</p>
		<p style="color:red; font-weight:bold;">Vui lòng đưa mã QR dưới cho nhân viên soát vé để vào rạp:</p>
		<img src="cid:ticket_qr" style="margin-top:10px;" alt="QR vé" />
		<p style="text-align:center; font-size:18px;"><strong>%s</strong></p>
	`, order.MovieName, order.TheaterName, order.BranchName,
		order.ShowDate, order.StartTime, seatHTML, foodHTML, order.Total, ticketCode)

	// Gửi email
	if err := services.SendInvoice(order.Email, subject, body, qrImage, "ticket_qr"); err != nil {
		return fmt.Errorf("send email failed: %v", err)
	}

	return nil
}

// func MomoResultHandler(c *gin.Context) {
// 	fmt.Println("MoMo IPN hit!") // hoặc dùng log.Println

// 	var momoResult struct {
// 		OrderId    string `json:"orderId"`
// 		RequestId  string `json:"requestId"`
// 		ResultCode int    `json:"resultCode"`
// 		Message    string `json:"message"`
// 		ExtraData  string `json:"extraData"`
// 		Signature  string `json:"signature"`
// 	}

// 	if err := c.ShouldBind(&momoResult); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MoMo response", "details": err.Error()})
// 		return
// 	}

// 	if momoResult.ResultCode != 0 {
// 		c.JSON(http.StatusOK, gin.H{"message": "Thanh toán thất bại hoặc bị huỷ"})
// 		return
// 	}

// 	// Giải mã extraData
// 	decodedBytes, err := base64.StdEncoding.DecodeString(momoResult.ExtraData)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode extraData"})
// 		return
// 	}

// 	var request struct {
// 		Order              models.Order       `json:"order"`
// 		OrderFoods         []models.OrderFood `json:"orderFoods"`
// 		ShowtimeSeatUpdate struct {
// 			ShowtimeSeatIDs []int `json:"ShowtimeSeatIDs"`
// 		} `json:"showtimeSeatUpdates"`
// 	}
// 	if err := json.Unmarshal(decodedBytes, &request); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse extraData"})
// 		return
// 	}

// 	// --- Ghi dữ liệu vào DB ---
// 	request.Order.CreatedAt = time.Now()
// 	if err := database.DB.Create(&request.Order).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
// 		return
// 	}

// 	for _, food := range request.OrderFoods {
// 		food.OrderID = request.Order.OrderID
// 		if err := database.DB.Create(&food).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order food"})
// 			return
// 		}
// 	}

// 	for _, seatID := range request.ShowtimeSeatUpdate.ShowtimeSeatIDs {
// 		var seat models.ShowtimeSeat
// 		if err := database.DB.First(&seat, seatID).Error; err != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Seat not found", "seatID": seatID})
// 			return
// 		}
// 		seat.Status = true
// 		if err := database.DB.Save(&seat).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update seat status"})
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Thanh toán thành công và đã lưu đơn hàng",
// 		"orderID": request.Order.OrderID,
// 	})
// }
