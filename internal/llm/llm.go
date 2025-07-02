package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ResultLLM struct {
	HasViolation bool   
	Category     string 
	Word         string 
	Context      string 
	IsSuspected  bool   // Подозрение, что это описание фильма
}

const promptTemplate = `
Ты - модератор, проверяющий текст на запрещённый контент: порнография, насилие, наркотики, спам и другое.

Если встречаются слова, которые могут быть связаны с названием или описанием фильмов/сериалов (например, "секс в большом городе"), не считай это нарушением, а пометь как "подозрение".

Выводи ответ в формате JSON с полями:
{
  "violation": true/false,
  "category": "<категория или None>",
  "word": "<найденное слово или пусто>",
  "context": "<контекст вокруг слова>",
  "suspected": true/false
}

Текст:
%s
`

func CheckWithLLM(text string) (ResultLLM, bool) {
	var result ResultLLM

	if strings.TrimSpace(text) == "" {
		return result, false
	}

	prompt := fmt.Sprintf(promptTemplate, text)

	
	reqBody := map[string]interface{}{
		"model": "llama2", 
		"prompt": prompt,
		"max_tokens": 500,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return result, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, false
	}

	var respData struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return result, false
	}
	if len(respData.Choices) == 0 {
		return result, false
	}

	llmOutput := respData.Choices[0].Text
	llmOutput = strings.TrimSpace(llmOutput)

	err = json.Unmarshal([]byte(llmOutput), &result)
	if err != nil {
		
		return result, false
	}

	return result, true
}
