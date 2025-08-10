package main

import (
	"log"
	"net/rpc"
	"encoding/json"
	sdk "github.com/orka-platform/orka-plugin-sdk"
)

func main() {
	// Connect to the LLM plugin
	client, err := rpc.Dial("tcp", "127.0.0.1:50052")
	if err != nil {
		log.Fatal("Failed to connect to plugin:", err)
	}
	defer client.Close()

	// Test 1: Basic OpenAI Chat Completion
	log.Println("Testing OpenAI Chat Completion...")
	
	openaiReq := sdk.Request{
		Method: "ChatCompletion",
		Args: map[string]any{
			"provider": "openai",
			"apiKey":   "your-openai-api-key-here", // Replace with actual API key
			"messages": []any{
				map[string]any{
					"role":    "user",
					"content": "What is the capital of France?",
				},
			},
		},
	}

	var openaiRes sdk.Response
	if err := client.Call("LLMPlugin.CallMethod", openaiReq, &openaiRes); err != nil {
		log.Printf("OpenAI call failed: %v", err)
	} else {
		log.Printf("OpenAI success: %v", openaiRes.Success)
		if openaiRes.Success {
			if data, ok := openaiRes.Data.(map[string]any); ok {
				if content, exists := data["content"]; exists {
					log.Printf("Response: %s", content)
				}
			}
		} else {
			log.Printf("Error: %s", openaiRes.Error)
		}
	}

	// Test 2: Anthropic Chat Completion with Custom Parameters
	log.Println("\nTesting Anthropic Chat Completion...")
	
	anthropicReq := sdk.Request{
		Method: "ChatCompletion",
		Args: map[string]any{
			"provider":   "anthropic",
			"apiKey":     "your-anthropic-api-key-here", // Replace with actual API key
			"model":      "claude-3-sonnet-20240229",
			"temperature": 0.5,
			"maxTokens":  500,
			"messages": []any{
				map[string]any{
					"role":    "system",
					"content": "You are a helpful assistant that gives concise answers.",
				},
				map[string]any{
					"role":    "user",
					"content": "Explain what is machine learning in one sentence.",
				},
			},
		},
	}

	var anthropicRes sdk.Response
	if err := client.Call("LLMPlugin.CallMethod", anthropicReq, &anthropicRes); err != nil {
		log.Printf("Anthropic call failed: %v", err)
	} else {
		log.Printf("Anthropic success: %v", anthropicRes.Success)
		if anthropicRes.Success {
			if data, ok := anthropicRes.Data.(map[string]any); ok {
				if content, exists := data["content"]; exists {
					log.Printf("Response: %s", content)
				}
				if model, exists := data["model"]; exists {
					log.Printf("Model used: %s", model)
				}
			}
		} else {
			log.Printf("Error: %s", anthropicRes.Error)
		}
	}

	// Test 3: Streaming Chat Completion
	log.Println("\nTesting Streaming Chat Completion...")
	
	streamReq := sdk.Request{
		Method: "StreamChatCompletion",
		Args: map[string]any{
			"provider": "openai",
			"apiKey":   "your-openai-api-key-here", // Replace with actual API key
			"messages": []any{
				map[string]any{
					"role":    "user",
					"content": "Write a haiku about programming.",
				},
			},
		},
	}

	var streamRes sdk.Response
	if err := client.Call("LLMPlugin.CallMethod", streamReq, &streamRes); err != nil {
		log.Printf("Streaming call failed: %v", err)
	} else {
		log.Printf("Streaming success: %v", streamRes.Success)
		if streamRes.Success {
			if data, ok := streamRes.Data.(map[string]any); ok {
				if content, exists := data["content"]; exists {
					log.Printf("Complete response: %s", content)
				}
				if chunks, exists := data["chunks"]; exists {
					if chunkArray, ok := chunks.([]any); ok {
						log.Printf("Received %d chunks", len(chunkArray))
					}
				}
			}
		} else {
			log.Printf("Error: %s", streamRes.Error)
		}
	}

	// Test 4: Error handling - Invalid provider
	log.Println("\nTesting Error Handling - Invalid Provider...")
	
	errorReq := sdk.Request{
		Method: "ChatCompletion",
		Args: map[string]any{
			"provider": "invalid-provider",
			"apiKey":   "dummy-key",
			"messages": []any{
				map[string]any{
					"role":    "user",
					"content": "This should fail",
				},
			},
		},
	}

	var errorRes sdk.Response
	if err := client.Call("LLMPlugin.CallMethod", errorReq, &errorRes); err != nil {
		log.Printf("Error call failed: %v", err)
	} else {
		log.Printf("Error test success: %v", errorRes.Success)
		if !errorRes.Success {
			log.Printf("Expected error: %s", errorRes.Error)
		}
	}

	log.Println("\nTest completed!")
}