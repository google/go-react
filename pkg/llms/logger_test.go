package llms_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-react/pkg/llms"
	llmstesting "github.com/google/go-react/pkg/llms/testing"
)

func TestLogger(t *testing.T) {
	t.Parallel()
	var fake llmstesting.Fake[int]
	var buf bytes.Buffer

	fake.Outputs = map[string]string{
		"some-prompt": "some-response",
	}

	logger := llms.NewLogger[int](&fake, &buf)

	resp, err := logger.Generate(context.Background(), "some-prompt", 1)
	if err != nil {
		t.Fatal(err)
	}

	if expected, actual := "some-response", resp; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
	if expected, actual := "some-prompt", fake.Prompts[0]; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
	if expected, actual := 1, fake.Params[0]; expected != actual {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if expected, actual := "some-prompt", m["prompt"]; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
	if expected, actual := 1.0, m["params"].(float64); expected != actual {
		t.Errorf("expected %f, got %f", expected, actual)
	}
	if expected, actual := "some-response", m["response"]; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestLogger_error(t *testing.T) {
	t.Parallel()
	var fake llmstesting.Fake[int]
	var buf bytes.Buffer

	fake.Err = errors.New("some-error")
	logger := llms.NewLogger[int](&fake, &buf)

	_, err := logger.Generate(context.Background(), "some-prompt", 1)
	if err == nil {
		t.Fatal("expected error")
	}

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if expected, actual := "some-prompt", m["prompt"]; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
	if expected, actual := 1.0, m["params"].(float64); expected != actual {
		t.Errorf("expected %f, got %f", expected, actual)
	}
	if expected, actual := "some-error", m["err"]; expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
