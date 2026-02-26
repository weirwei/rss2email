package conf

import (
	"os"

	"gopkg.in/yaml.v2"
)

var TranslationConf TranslationConfig

type TranslationConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Provider         string `yaml:"provider"`
	SourceLang       string `yaml:"source_lang"`
	TargetLang       string `yaml:"target_lang"`
	TimeoutSeconds   int    `yaml:"timeout_seconds"`
	OnlyTranslateEng bool   `yaml:"only_translate_english"`
	TranslateTitle   bool   `yaml:"translate_title"`
	TranslateContent bool   `yaml:"translate_content"`
	AppendTranslated bool   `yaml:"append_translated"`

	Google    GoogleTranslateConfig    `yaml:"google"`
	Microsoft MicrosoftTranslateConfig `yaml:"microsoft"`
	Baidu     BaiduTranslateConfig     `yaml:"baidu"`
	LLM       LLMTranslateConfig       `yaml:"llm"`
}

type GoogleTranslateConfig struct {
	APIKey string `yaml:"api_key"`
}

type MicrosoftTranslateConfig struct {
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	Region   string `yaml:"region"`
}

type BaiduTranslateConfig struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

type LLMTranslateConfig struct {
	Endpoint         string  `yaml:"endpoint"`
	APIKey           string  `yaml:"api_key"`
	Model            string  `yaml:"model"`
	Temperature      float64 `yaml:"temperature"`
	SystemPrompt     string  `yaml:"system_prompt"`
	AdditionalPrompt string  `yaml:"additional_prompt"`
}

func defaultTranslationConfig() TranslationConfig {
	return TranslationConfig{
		Enabled:          false,
		Provider:         "none",
		SourceLang:       "en",
		TargetLang:       "zh-CN",
		TimeoutSeconds:   15,
		OnlyTranslateEng: true,
		TranslateTitle:   true,
		TranslateContent: true,
		AppendTranslated: true,
		Microsoft: MicrosoftTranslateConfig{
			Endpoint: "https://api.cognitive.microsofttranslator.com",
		},
		LLM: LLMTranslateConfig{
			Endpoint: "https://api.openai.com/v1/chat/completions",
			Model:    "gpt-4o-mini",
			SystemPrompt: "You are a translation engine. Translate the user text from English to Simplified Chinese. " +
				"Only return the translated text. Keep URLs and markdown unchanged.",
		},
	}
}

func TranslationInit() {
	TranslationConf = defaultTranslationConfig()
	data, err := os.ReadFile("conf/yaml/translation.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	_ = yaml.Unmarshal(data, &TranslationConf)
}
