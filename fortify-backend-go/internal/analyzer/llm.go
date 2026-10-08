package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Providers:
//   - "ollama"    local model via Ollama (default, private)
//   - "gemini"    Google Gemini (cloud, BYOK)
//   - "openai"    OpenAI chat completions (cloud, BYOK); OPENAI_BASE_URL
//     override also serves any OpenAI-compatible endpoint
//   - "anthropic" Anthropic messages API (cloud, BYOK)
//   - "custom"    any OpenAI-compatible HTTP endpoint (BYOK + base URL + model:
//     Groq, Together, OpenRouter, Mistral, LM Studio, …)

var httpClient = &http.Client{Timeout: llmTimeout()}

// Full-report generation on a local CPU model can take minutes on a cold
// start; the default gives it room while LLM_TIMEOUT_SECONDS tunes it.
func llmTimeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv("LLM_TIMEOUT_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 600 * time.Second
}

// LLMOptions carries per-request provider settings. The dashboard sends these
// (BYOK); empty fields fall back to server-side env / built-in defaults.
type LLMOptions struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

func normalizeProvider(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	if p == "" {
		p = strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	}
	if p == "" {
		p = "ollama"
	}
	return p
}

// Defaults mirror the Python llm.py values, extended per provider. OLLAMA_URL
// defaults to localhost; docker-compose overrides it to
// host.docker.internal so the container reaches the host's Ollama.
func ollamaURL() string {
	if v := os.Getenv("OLLAMA_URL"); v != "" {
		return v
	}
	return "http://localhost:11434/api/generate"
}

func defaultModel(provider string) string {
	switch provider {
	case "gemini":
		if v := os.Getenv("GEMINI_MODEL"); v != "" {
			return v
		}
		return "gemini-3.6-flash"
	case "openai":
		if v := os.Getenv("OPENAI_MODEL"); v != "" {
			return v
		}
		return "gpt-4o-mini"
	case "anthropic":
		if v := os.Getenv("ANTHROPIC_MODEL"); v != "" {
			return v
		}
		return "claude-sonnet-4-5"
	case "custom":
		if v := os.Getenv("CUSTOM_LLM_MODEL"); v != "" {
			return v
		}
		return ""
	default: // ollama
		if v := os.Getenv("OLLAMA_MODEL"); v != "" {
			return v
		}
		return "llama3.1:8b"
	}
}

func geminiURL(model string) string {
	return "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent"
}

func openaiBaseURL() string {
	if v := strings.TrimRight(os.Getenv("OPENAI_BASE_URL"), "/"); v != "" {
		return v
	}
	return "https://api.openai.com/v1"
}

func customBaseURL(override string) string {
	if v := strings.TrimRight(strings.TrimSpace(override), "/"); v != "" {
		return v
	}
	return strings.TrimRight(os.Getenv("CUSTOM_LLM_BASE_URL"), "/")
}

const anthropicURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// loadDotEnv parses fortify-backend-go/.env (KEY=VALUE lines) for server-side
// defaults. The dashboard can still override per-request (BYOK); .env is only
// the fallback — same precedence as Python python-dotenv.
func loadDotEnv() {
	for _, candidate := range []string{".env", filepath.Join("fortify-backend-go", ".env")} {
		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			k, v, _ := strings.Cut(line, "=")
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(strings.Trim(strings.TrimSpace(v), `"'`))
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func init() { loadDotEnv() }

func effectiveModel(provider, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return defaultModel(provider)
}

// GetLLMResponse routes to the selected provider. Anything not matching a
// cloud provider falls back to Ollama (same "local by default" posture as
// the Python original).
func GetLLMResponse(prompt string, o LLMOptions) (string, error) {
	provider := normalizeProvider(o.Provider)
	model := effectiveModel(provider, o.Model)
	switch provider {
	case "gemini":
		key := o.APIKey
		if key == "" {
			key = os.Getenv("GEMINI_API_KEY")
		}
		return geminiResponse(prompt, model, key)
	case "openai":
		base := openaiBaseURL()
		if custom := strings.TrimRight(strings.TrimSpace(o.BaseURL), "/"); custom != "" {
			// A caller-supplied endpoint must never receive the server's key:
			// otherwise anyone could point base_url at their own server and
			// harvest OPENAI_API_KEY (and SSRF the backend). Require BYOK.
			if strings.TrimSpace(o.APIKey) == "" {
				return "", fmt.Errorf("a custom base URL requires your own API key (paste one in Counsel)")
			}
			base = custom
			return openAICompatibleResponse(base, o.APIKey, model, prompt, "OpenAI")
		}
		key := o.APIKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		return openAICompatibleResponse(base, key, model, prompt, "OpenAI")
	case "anthropic":
		key := o.APIKey
		if key == "" {
			key = os.Getenv("ANTHROPIC_API_KEY")
		}
		return anthropicResponse(prompt, model, key)
	case "custom":
		base := customBaseURL(o.BaseURL)
		if base == "" {
			return "", fmt.Errorf("custom provider needs a base URL (set it in Counsel or CUSTOM_LLM_BASE_URL)")
		}
		if model == "" {
			return "", fmt.Errorf("custom provider needs a model name (set it in Counsel or CUSTOM_LLM_MODEL)")
		}
		if fromRequest := strings.TrimRight(strings.TrimSpace(o.BaseURL), "/"); fromRequest != "" {
			// Same rule as OpenAI above: a caller-supplied endpoint only ever
			// gets a caller-supplied key — never the server's.
			if strings.TrimSpace(o.APIKey) == "" {
				return "", fmt.Errorf("a custom base URL requires your own API key (paste one in Counsel)")
			}
			return openAICompatibleResponse(base, o.APIKey, model, prompt, "custom endpoint")
		}
		key := o.APIKey
		if key == "" {
			key = os.Getenv("CUSTOM_LLM_API_KEY")
		}
		return openAICompatibleResponse(base, key, model, prompt, "custom endpoint")
	case "ollama":
		return ollamaResponse(prompt, model)
	default:
		return "", fmt.Errorf("unsupported provider %q (want ollama, gemini, openai, anthropic, or custom)", o.Provider)
	}
}

// KnownProvider reports whether p names a supported provider (empty means
// "server default", which is always valid).
func KnownProvider(p string) bool {
	if strings.TrimSpace(p) == "" {
		return true
	}
	switch normalizeProvider(p) {
	case "ollama", "gemini", "openai", "anthropic", "custom":
		return true
	}
	return false
}

// ResolveProviderModel reports the (provider, model) that WOULD handle the
// call, so every completed analysis is stamped — parity with Python.
func ResolveProviderModel(provider, modelOverride string) (string, string) {
	p := normalizeProvider(provider)
	return p, effectiveModel(p, modelOverride)
}

// extractJSON unmarshals strict JSON first, then falls back to the largest
// {...} span — some providers wrap JSON in prose despite instructions.
func extractJSON(raw string, v any) error {
	if err := json.Unmarshal([]byte(raw), v); err == nil {
		return nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("LLM returned invalid JSON")
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), v); err != nil {
		return fmt.Errorf("LLM returned invalid JSON")
	}
	return nil
}

func postJSON(url string, headers map[string]string, payload any, result any) error {
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("LLM endpoint returned non-JSON response (HTTP %d)", resp.StatusCode)
	}
	return nil
}

func ollamaResponse(prompt, model string) (string, error) {
	payload := map[string]any{
		"model":   model,
		"prompt":  prompt,
		"format":  "json",
		"stream":  false,
		"options": map[string]any{"temperature": 0, "seed": 42},
	}
	raw, _ := json.Marshal(payload)
	resp, err := httpClient.Post(ollamaURL(), "application/json", bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("ollama: %w — is Ollama running? Start it and pull %s", err, model)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	var data struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("ollama: invalid response")
	}
	if data.Error != "" {
		return "", fmt.Errorf("ollama: %s", data.Error)
	}
	return data.Response, nil
}

func geminiResponse(prompt, model, apiKey string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Gemini selected but no API key provided (paste one in Counsel, or set GEMINI_API_KEY)")
	}
	payload := map[string]any{
		"contents":         []any{map[string]any{"parts": []any{map[string]any{"text": prompt}}}},
		"generationConfig": map[string]any{"responseMimeType": "application/json", "temperature": 0},
	}
	var data struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := postJSON(geminiURL(model), map[string]string{"x-goog-api-key": apiKey}, payload, &data); err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}
	if data.Error.Message != "" || data.Error.Code != 0 {
		msg := strings.ToLower(data.Error.Message)
		switch {
		case data.Error.Code == 503 || strings.Contains(msg, "overloaded"):
			return "", fmt.Errorf("Gemini is temporarily overloaded (high demand). " +
				"Try again in a moment, or switch to Ollama in Settings.")
		case data.Error.Code == 429:
			return "", fmt.Errorf("Gemini rate limit reached — wait a bit and retry, or switch to Ollama.")
		default:
			return "", fmt.Errorf("Gemini error: %s", data.Error.Message)
		}
	}
	if len(data.Candidates) == 0 || len(data.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return data.Candidates[0].Content.Parts[0].Text, nil
}

// openAICompatibleResponse speaks the /chat/completions dialect with a JSON
// response format — used by OpenAI itself and every compatible endpoint
// (Groq, Together, OpenRouter, Mistral, LM Studio, …).
func openAICompatibleResponse(baseURL, apiKey, model, prompt, label string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("%s needs an API key (paste one in Counsel)", label)
	}
	payload := map[string]any{
		"model": model,
		"messages": []any{
			map[string]any{"role": "system", "content": "You are a senior application security analyst. Output a SINGLE JSON object and nothing else: no text before or after, no markdown code fences."},
			map[string]any{"role": "user", "content": prompt},
		},
		"temperature":     0,
		"response_format": map[string]any{"type": "json_object"},
	}
	var data struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	headers := map[string]string{"Authorization": "Bearer " + apiKey}
	if err := postJSON(baseURL+"/chat/completions", headers, payload, &data); err != nil {
		return "", fmt.Errorf("%s: %w", label, err)
	}
	if data.Error.Message != "" {
		return "", fmt.Errorf("%s error: %s", label, data.Error.Message)
	}
	if len(data.Choices) == 0 {
		return "", fmt.Errorf("%s: empty response", label)
	}
	return data.Choices[0].Message.Content, nil
}

func anthropicResponse(prompt, model, apiKey string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Anthropic selected but no API key provided (paste one in Counsel, or set ANTHROPIC_API_KEY)")
	}
	payload := map[string]any{
		"model":       model,
		"max_tokens":  4096,
		"temperature": 0,
		"messages": []any{
			map[string]any{"role": "user", "content": prompt},
		},
	}
	var data struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	headers := map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": anthropicVersion,
	}
	if err := postJSON(anthropicURL, headers, payload, &data); err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}
	if data.Error.Message != "" {
		return "", fmt.Errorf("Anthropic error: %s", data.Error.Message)
	}
	var sb strings.Builder
	for _, b := range data.Content {
		if b.Type == "text" {
			sb.WriteString(b.Text)
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return sb.String(), nil
}
