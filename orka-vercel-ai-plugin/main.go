package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sdk "github.com/orka-platform/orka-plugin-sdk"
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(map[string]string{})
	gob.Register([]string{})
}

type AIPlugin struct{}

type workerResponse struct {
	Success      bool              `json:"success"`
	Error        string            `json:"error,omitempty"`
	Text         string            `json:"text,omitempty"`
	Model        string            `json:"model,omitempty"`
	FinishReason string            `json:"finishReason,omitempty"`
	Usage        map[string]any    `json:"usage,omitempty"`
	Extra        map[string]any    `json:"extra,omitempty"`
}

func (p *AIPlugin) CallMethod(req sdk.Request, res *sdk.Response) error {
	switch req.Method {
	case "ChatCompletion":
		return p.handleChatCompletion(req, res)
	default:
		*res = sdk.Response{Success: false, Error: fmt.Sprintf("unknown method: %s", req.Method)}
		return nil
	}
}

func (p *AIPlugin) handleChatCompletion(req sdk.Request, res *sdk.Response) error {
	provider, _ := req.Args["provider"].(string)
	model, _ := req.Args["model"].(string)
	apiKey, _ := req.Args["apiKey"].(string)
	messages, _ := req.Args["messages"].([]any)

	if provider == "" || model == "" || apiKey == "" || len(messages) == 0 {
		*res = sdk.Response{Success: false, Error: "provider, model, apiKey and messages are required"}
		return nil
	}

	workerInput := map[string]any{}
	for k, v := range req.Args {
		workerInput[k] = v
	}

	wr, err := invokeNodeWorker(workerInput)
	if err != nil {
		*res = sdk.Response{Success: false, Error: err.Error()}
		return nil
	}

	if !wr.Success {
		*res = sdk.Response{Success: false, Error: wr.Error}
		return nil
	}

	data := map[string]any{
		"text":          wr.Text,
		"model":         wr.Model,
		"finishReason":  wr.FinishReason,
		"usage":         wr.Usage,
	}
	if wr.Extra != nil {
		for k, v := range wr.Extra {
			data[k] = v
		}
	}
	*res = sdk.Response{Success: true, Data: data}
	return nil
}

func invokeNodeWorker(payload map[string]any) (*workerResponse, error) {
	workerPath, err := resolveWorkerPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("node", workerPath)
	cmd.Env = sanitizedEnv()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe error: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe error: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start node worker: %w", err)
	}

	go func() {
		enc := json.NewEncoder(stdin)
		_ = enc.Encode(payload)
		_ = stdin.Close()
	}()

	// read stderr in background for diagnostics
	diagCh := make(chan string, 1)
	go func() {
		b := new(strings.Builder)
		r := bufio.NewReader(stderr)
		for {
			line, err := r.ReadString('\n')
			if len(line) > 0 { b.WriteString(line) }
			if err != nil {
				if !errors.Is(err, io.EOF) { b.WriteString("\n") }
				break
			}
		}
		diagCh <- strings.TrimSpace(b.String())
	}()

	var resp workerResponse
	dec := json.NewDecoder(stdout)
	if err := dec.Decode(&resp); err != nil {
		_ = cmd.Wait()
		diag := <-diagCh
		if diag != "" {
			return nil, fmt.Errorf("worker decode error: %v; stderr: %s", err, diag)
		}
		return nil, fmt.Errorf("worker decode error: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		diag := <-diagCh
		if diag != "" {
			return nil, fmt.Errorf("worker failed: %v; stderr: %s", err, diag)
		}
		return nil, fmt.Errorf("worker failed: %w", err)
	}

	return &resp, nil
}

func resolveWorkerPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine executable path: %w", err)
	}
	dir := filepath.Dir(exe)
	candidate := filepath.Join(dir, "worker", "index.mjs")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	// Fallback to repo layout when running from source
	alt := filepath.Join(".", "worker", "index.mjs")
	if _, err := os.Stat(alt); err == nil {
		return alt, nil
	}
	return "", fmt.Errorf("worker script not found: %s", candidate)
}

func sanitizedEnv() []string {
	// pass through a minimal environment; do not leak unrelated secrets
	env := []string{}
	for _, e := range os.Environ() {
		// drop common CI secrets patterns if present; keep PATH & NODE_OPTIONS if any
		if strings.HasPrefix(e, "AWS_") || strings.HasPrefix(e, "GOOGLE_") || strings.HasPrefix(e, "ANTHROPIC_") || strings.HasPrefix(e, "OPENAI_") {
			continue
		}
		env = append(env, e)
	}
	return env
}

func main() {
	port := flag.Int("port", 0, "TCP port for RPC server (required)")
	flag.Parse()
	if *port == 0 {
		fmt.Fprintln(os.Stderr, "Missing required --port argument")
		os.Exit(1)
	}

	if err := rpc.Register(&AIPlugin{}); err != nil {
		log.Fatalf("RPC register error: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	fmt.Printf("Vercel AI plugin listening on %s\n", addr)
	rpc.Accept(ln)
}