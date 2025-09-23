package controllers

import (
	"fmt"
	"movie-ticket-booking/chatbot"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatAIRequest struct {
	Messages []chatbot.ChatMessage `json:"messages"`
}

func ChatAIHandler(c *gin.Context) {
	var req ChatAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Truyền toàn bộ lịch sử vào AI
	reply, err := chatbot.AskAI(req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply": reply,
	})
}

type ClassicRequest struct {
	Action string `json:"action"`
	UserID int    `json:"userId"`
}

func ChatClassicHandler(c *gin.Context) {
	var req ClassicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Action == "" {
		c.JSON(http.StatusOK, gin.H{
			"reply": "Xin chào 👋, tôi có thể giúp gì cho bạn?",
			"menu": []string{
				"Phim đang chiếu?",
				"Phim sắp chiếu?",
				"Báo lỗi dịch vụ?",
				"Thông tin tài khoản?",
				"Lịch sử giao dịch?",
				"Diễn viên nổi bật?",
				"Đạo diễn nổi bật?",
				"Thông tin khuyến mãi?",
				"Blog điện ảnh?",
			},
		})
		return
	}

	switch req.Action {

	case "Thông tin tài khoản?":
		account, err := FindAccountByID(req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy dữ liệu tài khoản"})
			return
		}

		// ✅ Định dạng ngày sinh
		birthDateStr := ""
		if rawDate, ok := account["BirthDate"].(string); ok && rawDate != "" {
			if t, err := time.Parse("2006-01-02", rawDate); err == nil {
				birthDateStr = t.Format("02-01-2006")
			} else {
				birthDateStr = rawDate
			}
		}

		reply := fmt.Sprintf(
			"👤 Tên: %s\n📧 Email: %s\n📱 SĐT: %s\n🎂 Ngày sinh: %s",
			account["FullName"], account["Email"], account["PhoneNumber"], birthDateStr,
		)

		c.JSON(http.StatusOK, gin.H{
			"reply": reply,
		})

	case "Phim đang chiếu?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "🎬 Danh sách phim đang chiếu: \n- LÀM GIÀU VỚI MA 2: CUỘC CHIẾN HỘT XOÀN \n- EM XINH TINH QUÁI \n- CÔ DÂU MA \n- BĂNG ĐẢNG QUÁI KIỆT 2 \n- QUỶ MÓC RUỘT",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "/showing"},
			},
		})

	case "Phim sắp chiếu?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "🎬 Danh sách phim sắp chiếu: \n- KHẾ ƯỚC BÁN DÂU \n- TIỆM CẦM ĐỒ: CÓ CHƠI CÓ CHỊU \n- THE CONJURING: NGHI LỄ CUỐI CÙNG \n- ĐẠI CHIẾN XỨ SỞ CỐI XAY GIÓ",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "/coming"},
			},
		})

	case "Thông tin khuyến mãi?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "🔥 Các khuyến mãi hấp dẫn: \n- BACK TO SCHOOL \n- CHỤP ẢNH CÙNG PUI PUI \n- QUÀ TẶNG SINH NHẬT \n- Xem Phim Ngày Đôi \nTHIÊN LONG x DEMON SLAYER",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "https://www.cgv.vn/default/newsoffer"},
			},
		})

	case "Lịch sử giao dịch?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "📜 Lịch sử giao dịch gần nhất",
			"buttons": []map[string]string{
				{"label": "Tại đây ➜", "url": fmt.Sprintf("/history/%d", req.UserID)},
			},
		})

	case "Báo lỗi dịch vụ?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "⚠️ Báo lỗi dịch vụ",
			"buttons": []map[string]string{
				{"label": "Tại đây ➜", "url": "/chat"},
			},
		})

	case "Diễn viên nổi bật?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "🧝‍♂️ Diễn viên nổi bật hiện tại: \n- Chris Evans \n- Margot Robbie \n- Charlize Theron \n- Hugh Jackman \n- Robert Downey Jr. \n- Johnny Depp",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "https://www.galaxycine.vn/dien-vien/"},
			},
		})

	case "Đạo diễn nổi bật?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "🤵 Đạo diễn nổi bật hiện tại: \n- James Wan \n- Lê Bảo Trung \n- Đồng Đăng Giao \n- Khiếu Thú Dịch Tiểu Tinh",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "https://www.galaxycine.vn/dao-dien/"},
			},
		})

	case "Blog điện ảnh?":
		c.JSON(http.StatusOK, gin.H{
			"reply": "📰 Tin tức hot hiện tại:",
			"buttons": []map[string]string{
				{"label": "Xem chi tiết ➜", "url": "https://www.galaxycine.vn/movie-blog/"},
			},
		})

	default:
		c.JSON(http.StatusOK, gin.H{
			"reply": "Xin lỗi, tôi chưa hiểu lựa chọn này.",
		})
	}
}
