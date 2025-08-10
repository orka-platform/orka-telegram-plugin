package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"

	sdk "github.com/orka-platform/orka-plugin-sdk"
	"github.com/vercel/ai-sdk-go/ai"
	"github.com/vercel/ai-sdk-go/ai/anthropic"
	"github.com/vercel/ai-sdk-go/ai/openai"
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(map[string]string{})
	gob.Register([]string{})
}

type LLMPlugin struct{}

func (l *LLMPlugin) CallMethod(req sdk.Request, res *sdk.Response) error {
	switch req.Method {
	case "ChatCompletion":
		return l.handleChatCompletion(req, res)
	case "StreamChatCompletion":
		return l.handleStreamChatCompletion(req, res)
	default:
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("unknown method: %s", req.Method),
		}
		return nil
	}
}

func (l *LLMPlugin) handleChatCompletion(req sdk.Request, res *sdk.Response) error {
	// Extract arguments
	provider, _ := req.Args["provider"].(string)
	apiKey, _ := req.Args["apiKey"].(string)
	messages, _ := req.Args["messages"].([]any)
	model, _ := req.Args["model"].(string)
	temperature, _ := req.Args["temperature"].(float64)
	maxTokens, _ := req.Args["maxTokens"].(int)

	// Validate required arguments
	if provider == "" || apiKey == "" || len(messages) == 0 {
		*res = sdk.Response{
			Success: false,
			Error:   "provider, apiKey, and messages are required",
		}
		return nil
	}

	// Set defaults
	if model == "" {
		switch provider {
		case "openai":
			model = "gpt-3.5-turbo"
		case "anthropic":
			model = "claude-3-sonnet-20240229"
		default:
			*res = sdk.Response{
				Success: false,
				Error:   fmt.Sprintf("unsupported provider: %s", provider),
			}
			return nil
		}
	}

	if temperature == 0 {
		temperature = 0.7
	}

	if maxTokens == 0 {
		maxTokens = 1000
	}

	// Convert messages to AI SDK format
	aiMessages, err := l.convertMessages(messages)
	if err != nil {
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("invalid messages format: %v", err),
		}
		return nil
	}

	// Create context
	ctx := context.Background()

	// Handle different providers
	var completion *ai.Completion
	switch strings.ToLower(provider) {
	case "openai":
		completion, err = l.callOpenAI(ctx, apiKey, model, aiMessages, temperature, maxTokens)
	case "anthropic":
		completion, err = l.callAnthropic(ctx, apiKey, model, aiMessages, temperature, maxTokens)
	default:
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("unsupported provider: %s", provider),
		}
		return nil
	}

	if err != nil {
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("API call failed: %v", err),
		}
		return nil
	}

	*res = sdk.Response{
		Success: true,
		Data: map[string]any{
			"content":      completion.Content,
			"model":        completion.Model,
			"usage":        completion.Usage,
			"finishReason": completion.FinishReason,
		},
	}
	return nil
}

func (l *LLMPlugin) handleStreamChatCompletion(req sdk.Request, res *sdk.Response) error {
	// Extract arguments
	provider, _ := req.Args["provider"].(string)
	apiKey, _ := req.Args["apiKey"].(string)
	messages, _ := req.Args["messages"].([]any)
	model, _ := req.Args["model"].(string)
	temperature, _ := req.Args["temperature"].(float64)
	maxTokens, _ := req.Args["maxTokens"].(int)

	// Validate required arguments
	if provider == "" || apiKey == "" || len(messages) == 0 {
		*res = sdk.Response{
			Success: false,
			Error:   "provider, apiKey, and messages are required",
		}
		return nil
	}

	// Set defaults
	if model == "" {
		switch provider {
		case "openai":
			model = "gpt-3.5-turbo"
		case "anthropic":
			model = "claude-3-sonnet-20240229"
		default:
			*res = sdk.Response{
				Success: false,
				Error:   fmt.Sprintf("unsupported provider: %s", provider),
			}
			return nil
		}
	}

	if temperature == 0 {
		temperature = 0.7
	}

	if maxTokens == 0 {
		maxTokens = 1000
	}

	// Convert messages to AI SDK format
	aiMessages, err := l.convertMessages(messages)
	if err != nil {
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("invalid messages format: %v", err),
		}
		return nil
	}

	// Create context
	ctx := context.Background()

	// Handle different providers for streaming
	var stream *ai.Stream
	switch strings.ToLower(provider) {
	case "openai":
		stream, err = l.streamOpenAI(ctx, apiKey, model, aiMessages, temperature, maxTokens)
	case "anthropic":
		stream, err = l.streamAnthropic(ctx, apiKey, model, aiMessages, temperature, maxTokens)
	default:
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("unsupported provider: %s", provider),
		}
		return nil
	}

	if err != nil {
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("streaming API call failed: %v", err),
		}
		return nil
	}

	// Collect stream chunks
	var chunks []string
	for {
		chunk, err := stream.Next()
		if err != nil {
			break
		}
		if chunk.Content != "" {
			chunks = append(chunks, chunk.Content)
		}
	}

	*res = sdk.Response{
		Success: true,
		Data: map[string]any{
			"content": strings.Join(chunks, ""),
			"chunks":  chunks,
			"model":   model,
		},
	}
	return nil
}

func (l *LLMPlugin) convertMessages(messages []any) ([]ai.Message, error) {
	var aiMessages []ai.Message
	
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid message format")
		}
		
		role, _ := msgMap["role"].(string)
		content, _ := msgMap["content"].(string)
		
		if role == "" || content == "" {
			return nil, fmt.Errorf("role and content are required for each message")
		}
		
		aiMessages = append(aiMessages, ai.Message{
			Role:    role,
			Content: content,
		})
	}
	
	return aiMessages, nil
}

func (l *LLMPlugin) callOpenAI(ctx context.Context, apiKey, model string, messages []ai.Message, temperature float64, maxTokens int) (*ai.Completion, error) {
	client := openai.NewClient(apiKey)
	
	req := ai.CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	
	return client.Complete(ctx, req)
}

func (l *LLMPlugin) callAnthropic(ctx context.Context, apiKey, model string, messages []ai.Message, temperature float64, maxTokens int) (*ai.Completion, error) {
	client := anthropic.NewClient(apiKey)
	
	req := ai.CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	
	return client.Complete(ctx, req)
}

func (l *LLMPlugin) streamOpenAI(ctx context.Context, apiKey, model string, messages []ai.Message, temperature float64, maxTokens int) (*ai.Stream, error) {
	client := openai.NewClient(apiKey)
	
	req := ai.CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}
	
	return client.Stream(ctx, req)
}

func (l *LLMPlugin) streamAnthropic(ctx context.Context, apiKey, model string, messages []ai.Message, temperature float64, maxTokens int) (*ai.Stream, error) {
	client := anthropic.NewClient(apiKey)
	
	req := ai.CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}
	
	return client.Stream(ctx, req)
}

func main() {
	port := flag.Int("port", 0, "TCP port for RPC server (required)")
	flag.Parse()

	if *port == 0 {
		fmt.Fprintln(os.Stderr, "Missing required --port argument")
		os.Exit(1)
	}

	err := rpc.Register(&LLMPlugin{})
	if err != nil {
		log.Fatalf("RPC register error: %v", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	fmt.Printf("LLM plugin listening on %s\n", addr)
	rpc.Accept(listener)
}