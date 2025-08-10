# Orka LLM Plugin

This plugin provides chat completion functionality using various Large Language Model (LLM) providers through the Vercel AI SDK. It supports both regular chat completion and streaming responses.

## Features

- **Multiple LLM Providers**: Support for OpenAI and Anthropic
- **Chat Completion**: Standard chat completion with configurable parameters
- **Streaming**: Real-time streaming responses for better user experience
- **Flexible Configuration**: Customizable models, temperature, and token limits
- **Orka Integration**: Seamless integration with the Orka platform

## Supported Providers

### OpenAI
- **Default Model**: `gpt-3.5-turbo`
- **Supported Models**: Any OpenAI model (gpt-3.5-turbo, gpt-4, etc.)
- **API Endpoint**: Uses OpenAI's official API

### Anthropic
- **Default Model**: `claude-3-sonnet-20240229`
- **Supported Models**: Any Anthropic model (claude-3-sonnet, claude-3-opus, etc.)
- **API Endpoint**: Uses Anthropic's official API

## Installation

1. **Clone or download** this plugin to your Orka plugins directory
2. **Install dependencies**:
   ```bash
   cd orka-llm-plugin
   go mod tidy
   ```
3. **Build the plugin**:
   ```bash
   go build -o orka-llm-plugin
   ```

## Usage

### Starting the Plugin

```bash
./orka-llm-plugin --port 50052
```

The plugin will start listening on `127.0.0.1:50052` for RPC calls.

### Available Methods

#### 1. ChatCompletion

Sends a chat completion request and returns the full response.

**Arguments:**
- `provider` (string, required): LLM provider ("openai" or "anthropic")
- `apiKey` (string, required): API key for the specified provider
- `messages` (array, required): Array of message objects
- `model` (string, optional): Model name (uses defaults if not specified)
- `temperature` (number, optional): Sampling temperature (0.0-2.0, default: 0.7)
- `maxTokens` (number, optional): Maximum tokens to generate (default: 1000)

**Returns:**
- `content`: Generated text content
- `model`: Model used for completion
- `usage`: Token usage information
- `finishReason`: Reason for completion finish

#### 2. StreamChatCompletion

Sends a streaming chat completion request and returns both the complete content and individual chunks.

**Arguments:** Same as ChatCompletion

**Returns:**
- `content`: Complete generated text content
- `chunks`: Array of streaming chunks received
- `model`: Model used for completion

### Message Format

Messages should be provided as an array of objects with the following structure:

```json
[
  {
    "role": "system",
    "content": "You are a helpful assistant."
  },
  {
    "role": "user",
    "content": "Hello, how are you?"
  }
]
```

**Valid roles:**
- `system`: System instructions or context
- `user`: User input/messages
- `assistant`: Assistant responses (for conversation history)

## Examples

### Basic Chat Completion

```go
req := sdk.Request{
    Method: "ChatCompletion",
    Args: map[string]any{
        "provider": "openai",
        "apiKey":   "your-openai-api-key",
        "messages": []any{
            map[string]any{
                "role":    "user",
                "content": "What is the capital of France?",
            },
        },
    },
}
```

### Advanced Chat Completion with Custom Parameters

```go
req := sdk.Request{
    Method: "ChatCompletion",
    Args: map[string]any{
        "provider":   "anthropic",
        "apiKey":     "your-anthropic-api-key",
        "model":      "claude-3-opus-20240229",
        "temperature": 0.5,
        "maxTokens":  2000,
        "messages": []any{
            map[string]any{
                "role":    "system",
                "content": "You are a creative writing assistant.",
            },
            map[string]any{
                "role":    "user",
                "content": "Write a short story about a robot learning to paint.",
            },
        },
    },
}
```

### Streaming Chat Completion

```go
req := sdk.Request{
    Method: "StreamChatCompletion",
    Args: map[string]any{
        "provider": "openai",
        "apiKey":   "your-openai-api-key",
        "messages": []any{
            map[string]any{
                "role":    "user",
                "content": "Explain quantum computing in simple terms.",
            },
        },
    },
}
```

## Configuration

The plugin automatically sets sensible defaults:

- **Temperature**: 0.7 (balanced creativity and coherence)
- **Max Tokens**: 1000 (reasonable response length)
- **Models**: Provider-specific defaults (gpt-3.5-turbo for OpenAI, claude-3-sonnet for Anthropic)

## Error Handling

The plugin provides detailed error messages for common issues:

- Missing required arguments
- Invalid provider names
- API call failures
- Invalid message formats

## Security Considerations

- **API Keys**: Never log or expose API keys in responses
- **Local Binding**: Plugin binds only to localhost (127.0.0.1)
- **Input Validation**: All inputs are validated before processing

## Troubleshooting

### Common Issues

1. **"unknown method" error**: Ensure method names match exactly (case-sensitive)
2. **"unsupported provider" error**: Use "openai" or "anthropic" (case-insensitive)
3. **"API call failed" error**: Check API key validity and network connectivity
4. **"invalid messages format" error**: Ensure messages array contains valid objects with role and content

### Debug Mode

For debugging, you can add logging to the plugin by modifying the main.go file and rebuilding.

## Dependencies

- `github.com/orka-platform/orka-plugin-sdk`: Orka plugin SDK
- `github.com/vercel/ai-sdk-go`: Vercel AI SDK for Go

## License

This plugin follows the same license as the Orka platform.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

For issues and questions:
- Check the troubleshooting section
- Review the Orka plugin documentation
- Open an issue in the repository