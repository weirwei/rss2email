package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/conf"
)

type Translator interface {
	Translate(text string) (string, error)
}

type noopTranslator struct{}

func (t noopTranslator) Translate(text string) (string, error) {
	return text, nil
}

type httpTranslator struct {
	client *http.Client
	cfg    conf.TranslationConfig
}

func newTranslator(cfg conf.TranslationConfig) Translator {
	if !cfg.Enabled {
		ilog.Infof("translation disabled, use noop translator")
		return noopTranslator{}
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ilog.Infof(
		"translation enabled provider=%s source=%s target=%s timeout=%s only_english=%t title=%t content=%t append=%t",
		strings.ToLower(strings.TrimSpace(cfg.Provider)),
		cfg.SourceLang,
		cfg.TargetLang,
		timeout.String(),
		cfg.OnlyTranslateEng,
		cfg.TranslateTitle,
		cfg.TranslateContent,
		cfg.AppendTranslated,
	)
	return &httpTranslator{
		client: &http.Client{Timeout: timeout},
		cfg:    cfg,
	}
}

func (t *httpTranslator) Translate(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		ilog.Infof("translation skip: empty text")
		return text, nil
	}
	if t.cfg.OnlyTranslateEng && !looksEnglish(trimmed) {
		ilog.Infof("translation skip: text not detected as english, preview=%q", previewText(trimmed, 80))
		return text, nil
	}
	provider := strings.ToLower(strings.TrimSpace(t.cfg.Provider))
	ilog.Infof("translation request provider=%s chars=%d preview=%q", provider, len(trimmed), previewText(trimmed, 80))
	switch provider {
	case "google":
		return t.translateByGoogle(trimmed)
	case "microsoft", "azure":
		return t.translateByMicrosoft(trimmed)
	case "baidu":
		return t.translateByBaidu(trimmed)
	case "llm", "openai":
		return t.translateByLLM(trimmed)
	default:
		return text, nil
	}
}

func looksEnglish(text string) bool {
	var letters int
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			letters++
		}
		if letters >= 3 {
			return true
		}
	}
	return false
}

func (t *httpTranslator) translateByGoogle(text string) (string, error) {
	apiKey := strings.TrimSpace(t.cfg.Google.APIKey)
	if apiKey == "" {
		return text, nil
	}
	endpoint := "https://translation.googleapis.com/language/translate/v2"
	form := url.Values{}
	form.Set("q", text)
	form.Set("source", t.cfg.SourceLang)
	form.Set("target", t.cfg.TargetLang)
	form.Set("format", "text")
	form.Set("key", apiKey)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.client.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}
	if resp.StatusCode >= 300 {
		return text, fmt.Errorf("google translate http status: %d", resp.StatusCode)
	}
	var parsed struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}
	if err = json.Unmarshal(bodyBytes, &parsed); err != nil {
		return text, err
	}
	if len(parsed.Data.Translations) == 0 || strings.TrimSpace(parsed.Data.Translations[0].TranslatedText) == "" {
		return text, nil
	}
	return parsed.Data.Translations[0].TranslatedText, nil
}

func (t *httpTranslator) translateByMicrosoft(text string) (string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(t.cfg.Microsoft.Endpoint), "/")
	apiKey := strings.TrimSpace(t.cfg.Microsoft.APIKey)
	if endpoint == "" || apiKey == "" {
		return text, nil
	}
	targetLang := t.cfg.TargetLang
	if strings.EqualFold(targetLang, "zh-CN") {
		targetLang = "zh-Hans"
	}
	query := url.Values{}
	query.Set("api-version", "3.0")
	query.Set("from", t.cfg.SourceLang)
	query.Set("to", targetLang)
	uri := endpoint + "/translate?" + query.Encode()
	payload, _ := json.Marshal([]map[string]string{{"text": text}})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, uri, bytes.NewReader(payload))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", apiKey)
	if strings.TrimSpace(t.cfg.Microsoft.Region) != "" {
		req.Header.Set("Ocp-Apim-Subscription-Region", strings.TrimSpace(t.cfg.Microsoft.Region))
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}
	if resp.StatusCode >= 300 {
		return text, fmt.Errorf("microsoft translate http status: %d", resp.StatusCode)
	}
	var parsed []struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err = json.Unmarshal(bodyBytes, &parsed); err != nil {
		return text, err
	}
	if len(parsed) == 0 || len(parsed[0].Translations) == 0 || strings.TrimSpace(parsed[0].Translations[0].Text) == "" {
		return text, nil
	}
	return parsed[0].Translations[0].Text, nil
}

func (t *httpTranslator) translateByBaidu(text string) (string, error) {
	appID := strings.TrimSpace(t.cfg.Baidu.AppID)
	appSecret := strings.TrimSpace(t.cfg.Baidu.AppSecret)
	if appID == "" || appSecret == "" {
		return text, nil
	}
	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	raw := appID + text + salt + appSecret
	sum := md5.Sum([]byte(raw))
	sign := hex.EncodeToString(sum[:])
	form := url.Values{}
	form.Set("q", text)
	form.Set("from", t.cfg.SourceLang)
	form.Set("to", normalizeBaiduTargetLang(t.cfg.TargetLang))
	form.Set("appid", appID)
	form.Set("salt", salt)
	form.Set("sign", sign)
	endpoint := "https://fanyi-api.baidu.com/api/trans/vip/translate"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.client.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}
	if resp.StatusCode >= 300 {
		return text, fmt.Errorf("baidu translate http status: %d", resp.StatusCode)
	}
	var parsed struct {
		TransResult []struct {
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}
	if err = json.Unmarshal(bodyBytes, &parsed); err != nil {
		return text, err
	}
	if len(parsed.TransResult) == 0 || strings.TrimSpace(parsed.TransResult[0].Dst) == "" {
		return text, nil
	}
	return parsed.TransResult[0].Dst, nil
}

func (t *httpTranslator) translateByLLM(text string) (string, error) {
	// LLM translation uses single-shot streaming only:
	// no fallback to non-stream mode and no retry on failure.
	return t.translateByLLMStream(text)
}

func (t *httpTranslator) translateByLLMStream(text string) (string, error) {
	endpoint := strings.TrimSpace(t.cfg.LLM.Endpoint)
	apiKey := strings.TrimSpace(t.cfg.LLM.APIKey)
	model := strings.TrimSpace(t.cfg.LLM.Model)
	if endpoint == "" || apiKey == "" || model == "" {
		return text, nil
	}
	systemPrompt := strings.TrimSpace(t.cfg.LLM.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "You are a translation engine. Translate from English to Simplified Chinese only."
	}
	userText := text
	if strings.TrimSpace(t.cfg.LLM.AdditionalPrompt) != "" {
		userText = strings.TrimSpace(t.cfg.LLM.AdditionalPrompt) + "\n\n" + text
	}
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userText},
		},
		"temperature": t.cfg.LLM.Temperature,
		"stream":      true,
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := t.client.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if len(bodyBytes) > 0 {
			return text, fmt.Errorf("llm stream http status: %d body=%s", resp.StatusCode, previewText(string(bodyBytes), 160))
		}
		return text, fmt.Errorf("llm stream http status: %d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	var builder strings.Builder
	var eventData []string
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// SSE event terminator: blank line.
		if trimmed == "" {
			if len(eventData) == 0 {
				continue
			}
			raw := strings.TrimSpace(strings.Join(eventData, "\n"))
			eventData = eventData[:0]
			if raw == "" {
				continue
			}
			if raw == "[DONE]" {
				break
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			part := chunk.Choices[0].Delta.Content
			if strings.TrimSpace(part) == "" {
				part = chunk.Choices[0].Message.Content
			}
			if part != "" {
				builder.WriteString(part)
			}
			continue
		}

		// SSE comment line, ignore.
		if strings.HasPrefix(trimmed, ":") {
			continue
		}
		// We only care about `data:` lines and allow multi-line data payloads.
		if strings.HasPrefix(trimmed, "data:") {
			eventData = append(eventData, strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return text, err
	}
	if len(eventData) > 0 {
		raw := strings.TrimSpace(strings.Join(eventData, "\n"))
		if raw != "" && raw != "[DONE]" {
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(raw), &chunk); err == nil && len(chunk.Choices) > 0 {
				part := chunk.Choices[0].Delta.Content
				if strings.TrimSpace(part) == "" {
					part = chunk.Choices[0].Message.Content
				}
				if part != "" {
					builder.WriteString(part)
				}
			}
		}
	}
	translated := strings.TrimSpace(builder.String())
	if translated == "" {
		return text, fmt.Errorf("llm stream empty content")
	}
	return translated, nil
}

func normalizeBaiduTargetLang(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "zh-cn", "zh_hans", "zh-hans", "zh":
		return "zh"
	default:
		return target
	}
}

func safeTranslate(t Translator, text string) string {
	if t == nil || strings.TrimSpace(text) == "" {
		return text
	}
	translated, err := t.Translate(text)
	if err != nil {
		ilog.Warnf("translate failed: %v", err)
		return text
	}
	if strings.TrimSpace(translated) == "" {
		ilog.Warnf("translate returned empty content, fallback to source text")
		return text
	}
	if translated == text {
		ilog.Infof("translate unchanged, preview=%q", previewText(text, 80))
	} else {
		ilog.Infof("translate success src=%q dst=%q", previewText(text, 60), previewText(translated, 60))
	}
	return translated
}

func previewText(text string, max int) string {
	if max <= 0 {
		max = 80
	}
	compact := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	runes := []rune(compact)
	if len(runes) <= max {
		return compact
	}
	return string(runes[:max]) + "..."
}
