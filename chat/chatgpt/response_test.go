package chatgpt

import (
	"fmt"
	"testing"
)

func TestResponse(t *testing.T) {
	sampleJSON := `{
		"id": "chatcmpl-B4XSh7MGwzMXubIvkYMJVihDbQjcT",
		"object": "chat.completion",
		"created": 1740421639,
		"model": "gpt-4o-mini-2024-07-18",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "A billy bumbler is a fictional creature...",
					"refusal": null
				},
				"logprobs": null,
				"finish_reason": "stop"
			}
		],
		"usage": {
			"prompt_tokens": 16,
			"completion_tokens": 172,
			"total_tokens": 188,
			"prompt_tokens_details": {
				"cached_tokens": 0,
				"audio_tokens": 0
			},
			"completion_tokens_details": {
				"reasoning_tokens": 0,
				"audio_tokens": 0,
				"accepted_prediction_tokens": 0,
				"rejected_prediction_tokens": 0
			}
		},
		"service_tier": "default",
		"system_fingerprint": "fp_7fcd609668"
	}`

	parsedResponse, err := ParseResponse(sampleJSON)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	if parsedResponse.Object != "chat.completion" {
		fmt.Println("Object does not match")
	}
	if parsedResponse.Created != 1740421639 {
		fmt.Println("Created timestamp does not match")
	}
	if parsedResponse.Model != "gpt-4o-mini-2024-07-18" {
		fmt.Println("Model does not match")
	}
	if len(parsedResponse.Choices) != 1 || parsedResponse.Choices[0].Index != 0 {
		fmt.Println("Choices array does not match")
	}
	if parsedResponse.Choices[0].Message.Role != "assistant" || parsedResponse.Choices[0].Message.Content != "A billy bumbler is a fictional creature..." {
		fmt.Println("Message content does not match")
	}
	if parsedResponse.Usage.PromptTokens != 16 || parsedResponse.Usage.CompletionTokens != 172 || parsedResponse.Usage.TotalTokens != 188 {
		fmt.Println("Usage tokens do not match")
	}
	if parsedResponse.ServiceTier != "default" {
		fmt.Println("Service tier does not match")
	}
	if parsedResponse.SystemFingerprint != "fp_7fcd609668" {
		fmt.Println("System fingerprint does not match")
	}
}
