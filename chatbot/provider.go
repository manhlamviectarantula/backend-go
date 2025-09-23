package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"movie-ticket-booking/config"
	"net/http"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func AskAI(messages []ChatMessage) (string, error) {
	cfg := config.GetAIConfig()
	if cfg.APIKey == "" || cfg.BaseURL == "" {
		return "", fmt.Errorf("API key hoặc Base URL chưa được cấu hình")
	}

	// ✅ Thêm system prompt để giới hạn chủ đề
	systemPrompt := ChatMessage{
		Role: "system",
		Content: `Bạn là một trợ lý AI chuyên về phim ảnh. 
        - Chỉ trả lời các câu hỏi liên quan đến phim, rạp chiếu phim, lịch chiếu, diễn viên, đạo diễn, thể loại phim.  
        - Nếu người dùng hỏi về chính trị, xã hội, tôn giáo, hoặc chủ đề ngoài phạm vi phim ảnh thì hãy từ chối lịch sự bằng câu: 
        "Xin lỗi, tôi chỉ có thể hỗ trợ bạn về chủ đề phim ảnh và rạp chiếu."`,
	}

	// Ghép system prompt vào trước messages
	reqBody := ChatRequest{
		Model:    cfg.Model,
		Messages: append([]ChatMessage{systemPrompt}, messages...),
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(resBody))
	}

	var aiResp ChatResponse
	if err := json.Unmarshal(resBody, &aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiResp.Choices[0].Message.Content, nil
}
