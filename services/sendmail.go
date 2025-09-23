package services

import (
	"fmt"
	"movie-ticket-booking/config"
	"os"

	"gopkg.in/gomail.v2"
)

func SendMailReceiveTicket(to, code string) error {
	cfg := config.GetSendMailConfig()
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Xác thực email")
	m.SetBody("text/plain", fmt.Sprintf("Mã OTP: %s", code))

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Password)
	return d.DialAndSend(m)
}

func SendNewPasswordEmail(to, newPassword string) error {
	cfg := config.GetSendMailConfig()

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Khôi phục mật khẩu từ CINÉMÀ")
	m.SetBody("text/plain", fmt.Sprintf("Mật khẩu mới của bạn là: %s", newPassword))

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Password)
	return d.DialAndSend(m)
}

func SendInvoice(to, subject, body string, imageData []byte, cid string) error {
	cfg := config.GetSendMailConfig()

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Ghi ảnh QR ra file tạm
	tmpFile := "tmp_qr.png"
	if err := os.WriteFile(tmpFile, imageData, 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// Nhúng ảnh vào email với cid
	m.Embed(tmpFile)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Password)
	return d.DialAndSend(m)
}
