package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"

	sdk "github.com/orka-platform/orka-plugin-sdk"
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(map[string]string{})
	gob.Register([]string{})
}

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

// OrkaCall is the exported entrypoint symbol for in-process usage.
// It wraps the existing rpc-style method for minimal change.
func OrkaCall(req sdk.Request, res *sdk.Response) error {
	var t TelegramPlugin
	return t.CallMethod(req, res)
}
