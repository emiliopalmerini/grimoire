package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	reqID  atomic.Int64
	mu     sync.Mutex
}

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(lang *Language) (*Client, error) {
	cmd := exec.Command(lang.Command, lang.Args...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	return &Client{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

func (c *Client) Close() error {
	c.notify("shutdown", nil)
	c.notify("exit", nil)
	c.stdin.Close()
	return c.cmd.Wait()
}

func (c *Client) Initialize(ctx context.Context, rootPath string) error {
	absRoot, _ := filepath.Abs(rootPath)
	rootURI := "file://" + absRoot

	params := map[string]any{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"formatting": map[string]any{
					"dynamicRegistration": false,
				},
				"codeAction": map[string]any{
					"dynamicRegistration": false,
					"codeActionLiteralSupport": map[string]any{
						"codeActionKind": map[string]any{
							"valueSet": []string{
								"source.organizeImports",
							},
						},
					},
				},
			},
		},
	}

	_, err := c.call("initialize", params)
	if err != nil {
		return err
	}

	c.notify("initialized", map[string]any{})
	return nil
}

func (c *Client) OpenDocument(uri, languageID, content string) error {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": languageID,
			"version":    1,
			"text":       content,
		},
	}
	c.notify("textDocument/didOpen", params)
	return nil
}

func (c *Client) CloseDocument(uri string) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}
	c.notify("textDocument/didClose", params)
}

func (c *Client) Format(uri string) ([]TextEdit, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"options": map[string]any{
			"tabSize":      4,
			"insertSpaces": false,
		},
	}

	result, err := c.call("textDocument/formatting", params)
	if err != nil {
		return nil, err
	}

	var edits []TextEdit
	if err := json.Unmarshal(result, &edits); err != nil {
		return nil, fmt.Errorf("failed to parse formatting result: %w", err)
	}

	return edits, nil
}

func (c *Client) OrganizeImports(uri string, content string) ([]TextEdit, error) {
	lines := strings.Split(content, "\n")
	endLine := len(lines) - 1
	endChar := 0
	if endLine >= 0 && len(lines[endLine]) > 0 {
		endChar = len(lines[endLine])
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"range": map[string]any{
			"start": map[string]any{"line": 0, "character": 0},
			"end":   map[string]any{"line": endLine, "character": endChar},
		},
		"context": map[string]any{
			"diagnostics": []any{},
			"only":        []string{"source.organizeImports"},
		},
	}

	result, err := c.call("textDocument/codeAction", params)
	if err != nil {
		return nil, nil
	}

	var actions []CodeAction
	if err := json.Unmarshal(result, &actions); err != nil {
		return nil, nil
	}

	for _, action := range actions {
		if action.Kind == "source.organizeImports" && action.Edit != nil {
			for _, changes := range action.Edit.Changes {
				return changes, nil
			}
			for _, docEdit := range action.Edit.DocumentChanges {
				return docEdit.Edits, nil
			}
		}
	}

	return nil, nil
}

func (c *Client) DocumentSymbols(uri string) ([]DocumentSymbol, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}

	result, err := c.call("textDocument/documentSymbol", params)
	if err != nil {
		return nil, err
	}

	var symbols []DocumentSymbol

	var docSymbols []rawDocumentSymbol
	if err := json.Unmarshal(result, &docSymbols); err == nil && len(docSymbols) > 0 {
		var flatten func(syms []rawDocumentSymbol)
		flatten = func(syms []rawDocumentSymbol) {
			for _, s := range syms {
				kind := symbolKindNames[s.Kind]
				if kind == "" {
					kind = "Unknown"
				}
				symbols = append(symbols, DocumentSymbol{
					Name: s.Name,
					Kind: kind,
					Line: s.Range.Start.Line,
				})
				if len(s.Children) > 0 {
					flatten(s.Children)
				}
			}
		}
		flatten(docSymbols)
		return symbols, nil
	}

	var symInfos []rawSymbolInformation
	if err := json.Unmarshal(result, &symInfos); err == nil {
		for _, s := range symInfos {
			kind := symbolKindNames[s.Kind]
			if kind == "" {
				kind = "Unknown"
			}
			symbols = append(symbols, DocumentSymbol{
				Name: s.Name,
				Kind: kind,
				Line: s.Location.Range.Start.Line,
			})
		}
		return symbols, nil
	}

	return nil, fmt.Errorf("failed to parse document symbols")
}

func (c *Client) call(method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.reqID.Add(1)
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.send(req); err != nil {
		return nil, err
	}

	for {
		resp, err := c.receive()
		if err != nil {
			return nil, err
		}
		if resp.ID == id {
			if resp.Error != nil {
				return nil, fmt.Errorf("LSP error %d: %s", resp.Error.Code, resp.Error.Message)
			}
			return resp.Result, nil
		}
	}
}

func (c *Client) notify(method string, params any) error {
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	return c.send(req)
}

func (c *Client) send(req jsonrpcRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := c.stdin.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) receive() (*jsonrpcResponse, error) {
	var contentLength int

	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	data := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, data); err != nil {
		return nil, err
	}

	var resp jsonrpcResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
