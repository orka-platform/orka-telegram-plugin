## Orka Vercel AI Chat Plugin

This plugin exposes a `ChatCompletion` method over Orka's RPC contract, delegating to a Node worker that uses the Vercel AI SDK to call different providers (OpenAI, Anthropic, Google/Gemini).

### Methods

- ChatCompletion
  - args:
    - provider: `openai` | `anthropic` | `google`
    - model: model id (e.g. `gpt-4o-mini`, `claude-3-5-sonnet`, `gemini-1.5-pro`)
    - apiKey: provider API key
    - messages: array of `{ role, content }`
    - system: optional system prompt
    - temperature: optional number
    - maxTokens: optional number
    - baseURL: optional custom base URL
  - returns: `{ text, model, finishReason, usage }`

### Build

From this directory:

```bash
# build go plugin
go build -o orka-vercel-ai-plugin

# install worker deps
(cd worker && npm install)
```

### Run locally

```bash
./orka-vercel-ai-plugin --port 50061
```

Use Orka host to invoke `ChatCompletion` with the configured args and your provider API key.