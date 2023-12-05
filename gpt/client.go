package gpt

import (
	"context"
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/conversation"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/sashabaranov/go-openai"
)

type (
	Client interface {
		CreateChatCompletionStream(ctx context.Context, conv conversation.Conversation) (stream *openai.ChatCompletionStream, err error)
	}

	client struct {
		oc     *openai.Client
		model  string
		logger logger.Logger
	}
)

func (c *client) CreateChatCompletionStream(ctx context.Context, conv conversation.Conversation) (*openai.ChatCompletionStream, error) {
	stream, err := c.oc.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: conv.ToChatCompletionMessage(),
		},
	)

	if err != nil {
		c.logger.Log(logger.WARN, "failed to create chat completion stream, try fallback to gpt4: %v", err)

		// in 202311, gpt-4-1106-preview RPM may be extremely low. Try fallback only once.
		stream, err = c.oc.CreateChatCompletionStream(
			ctx,
			openai.ChatCompletionRequest{
				Model:    openai.GPT4,
				Messages: conv.ToChatCompletionMessage(),
			},
		)
	}

	return stream, err
}

func ProvideGPTClient(config config.Config, logger logger.Logger) Client {
	apiKey := config.OpenAIAPIKey()
	orgID := config.OpenAIOrganizationID()
	model := config.OpenAIModel()

	clientConfig := openai.DefaultConfig(apiKey)
	if orgID != "" {
		clientConfig.OrgID = orgID
	}

	return &client{
		oc:     openai.NewClientWithConfig(clientConfig),
		model:  model,
		logger: logger,
	}
}
