package llms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// NewLogger creates a logger that wraps the LLM. It will write both the
// prompt and response to the given io.Writer.
func NewLogger[TParams any](llm LLM[TParams], out io.Writer) LLM[TParams] {
	return logger[TParams]{
		llm: llm,
		out: out,
	}
}

type logger[TParams any] struct {
	llm LLM[TParams]
	out io.Writer
}

type loggerData[TParams any] struct {
	Prompt   string  `json:"prompt"`
	Response string  `json:"response"`
	Params   TParams `json:"params"`
	Err      string  `json:"err"`
}

// Generate implements the LLM interface.
func (l logger[TParams]) Generate(ctx context.Context, prompt string, params TParams) (out string, err error) {
	data := loggerData[TParams]{
		Prompt: prompt,
		Params: params,
	}
	defer func() {
		if e := json.NewEncoder(l.out).Encode(data); err != nil {
			err = fmt.Errorf("logger failed to encode and write to writer: %w", e)
			return
		}
	}()

	resp, err := l.llm.Generate(ctx, prompt, params)
	if err != nil {
		data.Err = err.Error()
		return "", err
	}
	data.Response = resp
	return resp, nil
}
