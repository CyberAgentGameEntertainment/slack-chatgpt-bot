package config

import (
	_ "embed"
	"github.com/SGE-AI/sge-bot/logger"
	"os"
	"strings"
)

//go:embed system.txt
var systemPrompt string

const (
	CustomInstructionsReplacement = "{{custom_instructions}}"
)

type (
	Config interface {
		SetBotUserID(botUserID string)
		BotUserID() string
		LogLevel() logger.LogLevel
		OpenAIAPIKey() string
		OpenAIOrganizationID() string
		SlackBotToken() string
		SlackAppLevelToken() string
		OpenAIModel() string
		SystemPrompt(customInstructions string) string
		GoogleApplicationCredentialsJSON() string
		GoogleServiceAccountEmail() string
		SpreadSheetID() string
	}

	config struct {
		loglevel             logger.LogLevel
		openAIAPIKey         string
		openAIOrganizationID string
		slackBotToken        string
		slackAppLevelToken   string
		openAIModel          string
		botUserID            string
	}
)

func (c *config) SetBotUserID(botUserID string) {
	c.botUserID = botUserID
}

func (c *config) BotUserID() string {
	return c.botUserID
}

func (c *config) SystemPrompt(customInstructions string) string {
	if customInstructions == "" {
		customInstructions = "no custom instructions"
	}

	return strings.Replace(systemPrompt, CustomInstructionsReplacement, customInstructions, 1)
}

func (c *config) LogLevel() logger.LogLevel {
	return c.loglevel
}

func (c *config) OpenAIAPIKey() string {
	return c.openAIAPIKey
}

func (c *config) OpenAIOrganizationID() string {
	return c.openAIOrganizationID
}

func (c *config) SlackBotToken() string {
	return c.slackBotToken
}

func (c *config) SlackAppLevelToken() string {
	return c.slackAppLevelToken
}

func (c *config) OpenAIModel() string {
	return c.openAIModel
}

func (c *config) GoogleApplicationCredentialsJSON() string {
	return os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")
}

func (c *config) GoogleServiceAccountEmail() string {
	return os.Getenv("GOOGLE_SERVICE_ACCOUNT_EMAIL")
}

func (c *config) SpreadSheetID() string {
	return os.Getenv("SPREADSHEET_ID")
}

func ProvideConfig() Config {
	logLevel := os.Getenv("LOG_LEVEL")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		panic("OPENAI_API_KEY is required")
	}

	openAIOrganizationID := os.Getenv("OPENAI_ORGANIZATION_ID")
	if openAIOrganizationID == "" {
		panic("OPENAI_ORGANIZATION_ID is required")
	}

	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackBotToken == "" {
		panic("SLACK_BOT_TOKEN is required")
	}

	slackAppLevelToken := os.Getenv("SLACK_APP_LEVEL_TOKEN")
	if slackAppLevelToken == "" {
		panic("SLACK_APP_LEVEL_TOKEN is required")
	}

	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-4"
	}

	return &config{
		loglevel:             logger.LogLevel(logLevel),
		openAIAPIKey:         openAIAPIKey,
		openAIOrganizationID: openAIOrganizationID,
		slackBotToken:        slackBotToken,
		slackAppLevelToken:   slackAppLevelToken,
		openAIModel:          openAIModel,
	}
}
