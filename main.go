package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"

	sdk "github.com/orka-platform/orka-plugin-sdk"
)

type TelegramPlugin struct{}

func (t *TelegramPlugin) CallMethod(req sdk.Request, res *sdk.Response) error {
	switch req.Method {
	case "SendMessage":
		token, _ := req.Args["token"].(string)
		chatID, _ := req.Args["chatID"].(string)
		text, _ := req.Args["text"].(string)

		if token == "" || chatID == "" || text == "" {
			*res = sdk.Response{
				Success: false,
				Error:   "token, chatID and text are required",
			}
			return nil
		}

		err := sendTelegramMessage(token, chatID, text)
		if err != nil {
			*res = sdk.Response{Success: false, Error: err.Error()}
		} else {
			*res = sdk.Response{Success: true, Data: map[string]any{"messageID": "123"}}
		}
		return nil

	default:
		*res = sdk.Response{
			Success: false,
			Error:   fmt.Sprintf("unknown method: %s", req.Method),
		}
		return nil
	}
}

func sendTelegramMessage(token, chatID, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", text)

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %s", resp.Status)
	}

	return nil
}

func main() {
	port := flag.Int("port", 0, "TCP port for RPC server (required)")
	flag.Parse()

	if *port == 0 {
		fmt.Fprintln(os.Stderr, "Missing required --port argument")
		os.Exit(1)
	}

	err := rpc.Register(&TelegramPlugin{})
	if err != nil {
		log.Fatalf("RPC register error: %v", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	fmt.Printf("Telegram plugin listening on %s\n", addr)
	rpc.Accept(listener)
}
