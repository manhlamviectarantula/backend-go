package services

import (
	"encoding/base64"
	"fmt"
	"movie-ticket-booking/config"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Gửi OTP
func SendMailReceiveTicket(to, code string) error {
	cfg := config.GetSendMailConfig()

	from := mail.NewEmail("", cfg.From)
	subject := "Xác thực email nhận vé từ CINEMA"
	toEmail := mail.NewEmail("", to)
	body := fmt.Sprintf("Mã OTP của bạn là: %s", code)

	message := mail.NewSingleEmail(from, subject, toEmail, body, body)
	client := sendgrid.NewSendClient(cfg.APIKey)
	_, err := client.Send(message)
	return err
}

// Gửi mật khẩu mới
func SendNewPasswordEmail(to, newPassword string) error {
	cfg := config.GetSendMailConfig()

	from := mail.NewEmail("", cfg.From)
	subject := "Khôi phục mật khẩu CINEMA"
	toEmail := mail.NewEmail("", to)
	body := fmt.Sprintf("Mật khẩu mới của bạn là: %s", newPassword)

	message := mail.NewSingleEmail(from, subject, toEmail, body, body)
	client := sendgrid.NewSendClient(cfg.APIKey)
	_, err := client.Send(message)
	return err
}

// Gửi hóa đơn / invoice kèm QR code inline
func SendInvoice(to, subject, body string, imageData []byte, cid string) error {
	cfg := config.GetSendMailConfig()

	from := mail.NewEmail("", cfg.From)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	// Thêm inline attachment (QR code)
	attachment := mail.NewAttachment()
	encoded := base64.StdEncoding.EncodeToString(imageData)
	attachment.SetContent(encoded)
	attachment.SetType("image/png")
	attachment.SetFilename("qr.png")
	attachment.SetDisposition("inline")
	attachment.SetContentID(cid)
	message.AddAttachment(attachment)

	client := sendgrid.NewSendClient(cfg.APIKey)
	_, err := client.Send(message)
	return err
}
