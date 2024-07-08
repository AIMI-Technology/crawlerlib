package classifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ClassifierRequest struct {
	Data string `json:"data"`
}

type ClassifierResponse struct {
	Resp string `json:"resp"`
}

func IsArticleRelevant(inputText string) bool {

	requestData := ClassifierRequest{
		Data: inputText,
	}

	// Marshal the requestData to JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}

	classifierUrl := os.Getenv("CLASSIFIER_URL")
	if classifierUrl == "" {
		classifierUrl = "http://178.128.134.67"
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/classify", classifierUrl), bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Unmarshal the JSON response into the OpenAIResponse struct
	var response ClassifierResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		panic(err)
	}
	// Check if there is at least one choice and print the content
	if len(response.Resp) == 0 {
		panic(errors.New("no response from GPT"))
	}

	return strings.ToLower(response.Resp) == "positive"
}
