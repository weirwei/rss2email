package service

import (
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
		return noopTranslator{}
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &httpTranslator{
		client: &http.Client{Timeout: timeout},
		cfg:    cfg,
	}
}

func (t *httpTranslator) Translate(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text, nil
	}
	if t.cfg.OnlyTranslateEng && !looksEnglish(trimmed) {
		return text, nil
	}
	switch strings.ToLower(strings.TrimSpace(t.cfg.Provider)) {
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
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		return text, fmt.Errorf("llm translate http status: %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.Unmarshal(bodyBytes, &parsed); err != nil {
		return text, err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return text, nil
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
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
		return text
	}
	return translated
}
