package utils

import (
	"bytes"
	crand "crypto/rand"
	"image/png"
	mrand "math/rand"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func GenOTP(n int) string {
	const chars = "0123456789"
	b := make([]byte, n)
	crand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func GenerateRandomPassword(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	seededRand := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	password := make([]byte, length)
	for i := range password {
		password[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(password)
}

func GenerateTicketCode(n int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	mrand.Seed(time.Now().UnixNano())

	code := make([]byte, n)
	for i := range code {
		code[i] = charset[mrand.Intn(len(charset))]
	}
	return string(code)
}

func GenerateQRCode(code string) ([]byte, error) {
	qrCode, err := qr.Encode(code, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}
	qrCode, _ = barcode.Scale(qrCode, 256, 256)

	var buf bytes.Buffer
	if err := png.Encode(&buf, qrCode); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
