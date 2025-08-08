## Building an Orka Plugin (Example: Telegram Messenger)

This repository demonstrates a minimal Orka plugin implemented in Go that sends messages via Telegram. Use it as a reference or starting point for building your own plugins.

### What is an Orka plugin?

An Orka plugin is a small process the Orka host starts with a TCP `--port` argument. The plugin starts a Go `net/rpc` server and exposes a single entrypoint method that Orka calls with a generic request:

- **Request**: contains the target `Method` and a map of `Args`
- **Response**: indicates `Success`, optional `Error`, and optional `Data`

This contract is provided by `github.com/orka-platform/orka-plugin-sdk`.

---

### Repository structure

- `main.go`: Plugin implementation and RPC server bootstrap
- `config.json`: Metadata and method specification for the Orka host (UI/registry)
- `go.mod`, `go.sum`: Go module definition and dependencies

---

### Prerequisites

- Go 1.23+
- A Telegram Bot token (for testing this specific example)

---

### How this example works

The plugin defines a type with a `CallMethod` function and registers it with Go `net/rpc`:

```go
type TelegramPlugin struct{}

func (t *TelegramPlugin) CallMethod(req sdk.Request, res *sdk.Response) error {
    switch req.Method {
    case "SendMessage":
        token, _ := req.Args["token"].(string)
        chatID, _ := req.Args["chatID"].(string)
        text, _ := req.Args["text"].(string)
        // validate → call external API → fill response
    default:
        // unknown method
    }
}
```

At startup, the plugin:

1. Parses the required `--port` argument
2. Registers the plugin object with `rpc.Register`
3. Listens on `127.0.0.1:<port>` and accepts RPC calls

The `SendMessage` method calls Telegram's `sendMessage` HTTP API with the provided `token`, `chatID`, and `text`.

---

### The `config.json` contract

`config.json` describes your plugin for Orka's registry/UI and must mirror what your code expects.

- **name**: Human-readable plugin name
- **version**: Semantic version of your plugin
- **description**: Short description
- **tags**: Array of tags
- **methods**: A dictionary of exposed methods with their argument schemas

Important: Method names and argument keys are case-sensitive and must match what your code expects in `req.Method` and `req.Args`. For this example, the code expects method `SendMessage` and args `token`, `chatID`, and `text`.

Example (aligned with the code):

```json
{
  "name": "Telegram",
  "version": "0.0.0",
  "description": "Telegram plugin to send messages",
  "tags": ["telegram", "messenger", "social-media"],
  "methods": {
    "SendMessage": {
      "description": "Sends a message to a chat",
      "args": [
        { "name": "token",  "description": "Bot auth token", "type": "string" },
        { "name": "chatID", "description": "Chat id to send the message to", "type": "string" },
        { "name": "text",   "description": "Message text", "type": "string" }
      ]
    }
  }
}
```

---

### Run locally

Build and run the plugin (it listens on TCP RPC, not HTTP):

```bash
go build -o orka-telegram-plugin
./orka-telegram-plugin --port 50051
```

Expected log:

```
Telegram plugin listening on 127.0.0.1:50051
```

To test end-to-end, let the Orka host launch this plugin and invoke `SendMessage` with the correct args. If you need a direct test, you can write a small Go RPC client using the same `sdk.Request`/`sdk.Response` types and call `CallMethod` over TCP.

Minimal example client (for local testing only):

```go
package main

import (
    "log"
    "net/rpc"
    sdk "github.com/orka-platform/orka-plugin-sdk"
)

func main() {
    client, err := rpc.Dial("tcp", "127.0.0.1:50051")
    if err != nil { log.Fatal(err) }
    defer client.Close()

    req := sdk.Request{
        Method: "SendMessage",
        Args: map[string]any{
            "token":  "<TELEGRAM_BOT_TOKEN>",
            "chatID": "<CHAT_ID>",
            "text":   "Hello from Orka plugin!",
        },
    }

    var res sdk.Response
    if err := client.Call("TelegramPlugin.CallMethod", req, &res); err != nil {
        log.Fatal(err)
    }
    log.Printf("success=%v error=%q data=%v", res.Success, res.Error, res.Data)
}
```

---

### Create your own plugin (step-by-step)

1) Initialize a new module

```bash
mkdir orka-my-plugin && cd orka-my-plugin
go mod init github.com/your-org/orka-my-plugin
go get github.com/orka-platform/orka-plugin-sdk@latest
```

2) Implement `main.go`

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "net/rpc"
    "os"
    sdk "github.com/orka-platform/orka-plugin-sdk"
)

type MyPlugin struct{}

func (p *MyPlugin) CallMethod(req sdk.Request, res *sdk.Response) error {
    switch req.Method {
    case "MyMethod":
        // 1) read args from req.Args
        // 2) validate inputs
        // 3) do the work (call external API, DB, etc.)
        // 4) fill response
        *res = sdk.Response{Success: true, Data: "OK"}
        return nil
    default:
        *res = sdk.Response{Success: false, Error: fmt.Sprintf("unknown method: %s", req.Method)}
        return nil
    }
}

func main() {
    port := flag.Int("port", 0, "TCP port for RPC server (required)")
    flag.Parse()
    if *port == 0 { fmt.Fprintln(os.Stderr, "Missing required --port argument"); os.Exit(1) }

    if err := rpc.Register(&MyPlugin{}); err != nil { log.Fatalf("RPC register error: %v", err) }
    addr := fmt.Sprintf("127.0.0.1:%d", *port)
    ln, err := net.Listen("tcp", addr)
    if err != nil { log.Fatalf("Failed to listen on %s: %v", addr, err) }
    fmt.Printf("My plugin listening on %s\n", addr)
    rpc.Accept(ln)
}
```

3) Define `config.json`

Ensure names match the code exactly (case-sensitive):

```json
{
  "name": "My Plugin",
  "version": "0.1.0",
  "description": "Does something useful",
  "tags": ["sample"],
  "methods": {
    "MyMethod": {
      "description": "What this method does",
      "args": [
        { "name": "foo", "type": "string", "description": "Required input" }
      ]
    }
  }
}
```

4) Build and run

```bash
go build -o orka-my-plugin
./orka-my-plugin --port 50051
```

5) Integrate with Orka

- Place the built binary where Orka can execute it, or register it in your Orka configuration
- Ensure Orka passes the `--port` argument when launching the plugin
- Orka will call `CallMethod` with `sdk.Request` that matches your `config.json`

---

### Best practices

- **Match names exactly**: `req.Method` and `req.Args` keys must align 1:1 with `config.json`
- **Validate inputs** early and return actionable errors via `sdk.Response.Error`
- **Handle external API failures** robustly; include HTTP status codes and error messages
- **Never log secrets** (e.g., tokens); prefer redaction
- **Bind to localhost** (as shown) to avoid exposing your RPC port broadly
- **Use semantic versioning** in `config.json.version`

---

### Troubleshooting

- Missing `--port`: the process will exit; pass a non-zero port
- Connection refused: ensure the plugin is running and bound to `127.0.0.1:<port>`
- Method not found: check case-sensitive method name in `config.json` vs `req.Method`
- Argument missing/wrong type: ensure the caller sends exactly the keys and types your code expects

---

### License

Use your preferred license for your plugin. If contributing back, follow your organization's policies.


