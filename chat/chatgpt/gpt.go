package chatgpt

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const apiURL = "https://api.openai.com/v1/chat/completions"

// RequestBody represents the JSON payload structure
type RequestBody struct {
	Model    string    `json:"model"`
	Store    bool      `json:"store"`
	Messages []Message `json:"messages"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Refusal *string `json:"refusal"`
}

// Completer is the main struct that holds the configuration for the OpenAI API.
type Completer struct {
	Model
	Store  bool
	Role   string
	APIKey string

	client http.Client
}

// Model represents the model to use for completion.
type Model string

const (
	// ModelGPT4oMini is the default model used if Model is zeroed out.
	ModelGPT4oMini = Model("gpt-4o-mini")

	// DefaultRole is the default role used if Role is zeroed out.
	DefaultRole = "user"
)

// Complete sends a message to the OpenAI API and returns the response.
func (c Completer) Complete(message string) (string, error) {
	if c.Model == "" {
		c.Model = ModelGPT4oMini
	}

	if c.Role == "" {
		c.Role = DefaultRole
	}

	if c.APIKey == "" {
		return "", errors.New("Blank API key is forbidden")
	}

	// Construct the request body
	requestBody := RequestBody{
		Model: string(c.Model),
		Store: c.Store,
		Messages: []Message{
			{Role: c.Role, Content: message},
		},
	}

	// Convert the struct to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "marshalling JSON")
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.Wrap(err, "creating request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Create HTTP client and send request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "sending request")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading response body")
	}

	// Parse response
	response, err := ParseResponse(string(body))
	if err != nil {
		return "", errors.Wrap(err, "parsing response")
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}
