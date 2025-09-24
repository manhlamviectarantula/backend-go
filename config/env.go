package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Không tìm thấy file .env, dùng biến môi trường hệ thống nếu có")
	}
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetJWTKey() []byte {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		log.Fatal("JWT_KEY not set in environment")
	}
	return []byte(key)
}

func GetFacebookConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     GetEnv("FACEBOOK_APP_ID", ""),
		ClientSecret: GetEnv("FACEBOOK_APP_SECRET_KEY", ""),
		RedirectURL:  GetEnv("FACEBOOK_REDIRECT_URL", ""),
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}

func GetGoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     GetEnv("GOOGLE_APP_ID", ""),
		ClientSecret: GetEnv("GOOGLE_APP_SECRET_KEY", ""),
		RedirectURL:  GetEnv("GOOGLE_REDIRECT_URL", ""),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// type SendMailConfig struct {
// 	From     string
// 	Password string
// 	Host     string
// 	Port     int
// }

//	func GetSendMailConfig() *SendMailConfig {
//		port, _ := strconv.Atoi(GetEnv("EMAIL_SMTP_PORT", "587"))
//		return &SendMailConfig{
//			From:     GetEnv("EMAIL_FROM", ""),
//			Password: GetEnv("EMAIL_PASSWORD", ""),
//			Host:     GetEnv("EMAIL_SMTP_HOST", "smtp.gmail.com"),
//			Port:     port,
//		}
//	}
//

// cấu hình gửi mail cho SendGrid
type SendMailConfig struct {
	From   string
	APIKey string
}

// GetSendMailConfig đọc từ .env
func GetSendMailConfig() *SendMailConfig {
	return &SendMailConfig{
		From:   GetEnv("EMAIL_FROM", ""),
		APIKey: GetEnv("SENDGRID_API_KEY", ""),
	}
}

func GetMomoEnv() map[string]string {
	return map[string]string{
		"PARTNER_CODE": GetEnv("MOMO_PARTNER_CODE", ""),
		"ACCESS_KEY":   GetEnv("MOMO_ACCESS_KEY", ""),
		"SECRET_KEY":   GetEnv("MOMO_SECRET_KEY", ""),
		"REDIRECT_URL": GetEnv("MOMO_REDIRECT_URL", ""),
		"IPN_URL":      GetEnv("MOMO_IPN_URL", ""),
	}
}

type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func GetAIConfig() *AIConfig {
	return &AIConfig{
		APIKey:  GetEnv("AI_API_KEY", ""),
		BaseURL: GetEnv("AI_BASE_URL", ""),
		Model:   GetEnv("AI_MODEL", ""),
	}
}
