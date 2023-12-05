package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/conversation"
	"github.com/SGE-AI/sge-bot/gpt"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/SGE-AI/sge-bot/repository"
	"github.com/SGE-AI/sge-bot/slackapi"
	"github.com/sashabaranov/go-openai"
	"io"
	"time"
)

const (
	AckMessage = "..."

	UpdateInterval = 3 * time.Second

	UpdatingMessage = "..."

	OnErrorMessage = "APIの呼び出しでエラーが発生しました。しばらく時間をおいてから、もう一度お試しください。"
)

type (
	Chat interface {
		// StartNormalConversation - 通常の会話を開始します
		StartNormalConversation(channelID string, threadTS string) error

		// RegenerateMessage - 指定したoutputTSの会話を再生成します
		RegenerateMessage(channelID string, outputTS string, threadTS string, controllerTS string) error

		// StopGenerateMessage - 指定したoutputTSの会話の生成を停止します
		StopGenerateMessage(channelID string, outputTS string, controllerTS string) error

		// DeleteMessage - 指定したoutputTSの会話を削除します
		DeleteMessage(channelID string, outputTS string, controllerTS string) error
	}

	chat struct {
		slack  slackapi.SlackAPI
		gpt    gpt.Client
		config config.Config
		logger logger.Logger
		crepo  repository.ContextCancelRepository
	}
)

func (c chat) StartNormalConversation(channelID string, threadTS string) error {
	botMessage, err := c.slack.CreateNewBotMessage(channelID, threadTS, AckMessage)
	if err != nil {
		return fmt.Errorf("failed to fast post ack message: %v", err)
	}

	ci, err := c.slack.LoadCustomInstructions(channelID)
	if err != nil {
		c.logger.Log(logger.WARN, "failed to load conversation topic: %v", err)
	}

	messages, err := c.slack.LoadConversationReplies(channelID, threadTS)
	if err != nil {
		_ = botMessage.UpdateMessage(OnErrorMessage, false)
		return fmt.Errorf("failed to load conversation replies: %v", err)
	}

	conv := conversation.NewConversationFromSlackMessages(messages, c.config.BotUserID())
	conv.SystemMessage(c.config.SystemPrompt(ci))
	conv.RemoveMessageAfterTimestamp(botMessage.OutputTimeStamp())

	return c.startConversation(botMessage, conv)
}

func (c chat) RegenerateMessage(channelID string, outputTS string, threadTS string, controllerTS string) error {
	cancel, ok := c.crepo.Load(outputTS)
	if ok {
		cancel()
	}

	botMessage, err := c.slack.TakeOverBotMessage(channelID, outputTS, controllerTS)
	if err != nil {
		return fmt.Errorf("failed to take over bot message: %v", err)
	}

	botMessage.Regenerate(AckMessage)

	ci, err := c.slack.LoadCustomInstructions(channelID)
	if err != nil {
		c.logger.Log(logger.WARN, "failed to load conversation topic: %v", err)
	}

	messages, err := c.slack.LoadConversationReplies(channelID, threadTS)
	if err != nil {
		_ = botMessage.UpdateMessage(OnErrorMessage, false)
		return fmt.Errorf("failed to load conversation replies: %v", err)
	}

	conv := conversation.NewConversationFromSlackMessages(messages, c.config.BotUserID())
	conv.SystemMessage(c.config.SystemPrompt(ci))
	conv.RemoveMessageAfterTimestamp(botMessage.OutputTimeStamp())

	return c.startConversation(botMessage, conv)
}

func (c chat) DeleteMessage(channelID string, outputTS string, controllerTS string) error {
	cancel, ok := c.crepo.Load(outputTS)
	if ok {
		cancel()
	}

	botMessage, err := c.slack.TakeOverBotMessage(channelID, outputTS, controllerTS)
	if err != nil {
		return fmt.Errorf("failed to take over bot message: %v", err)
	}

	return botMessage.DeleteMySelf()
}

func (c chat) StopGenerateMessage(channelID string, outputTS string, controllerTS string) error {
	cancel, ok := c.crepo.Load(outputTS)
	if ok {
		cancel()
	}

	return nil
}

func (c chat) startConversation(botMessage slackapi.BotMessage, conv conversation.Conversation) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := c.crepo.Save(botMessage.OutputTimeStamp(), cancel)
	if err != nil {
		return fmt.Errorf("failed to save context cancel: %v", err)
	}
	defer c.crepo.Delete(botMessage.OutputTimeStamp())

	stream, err := c.gpt.CreateChatCompletionStream(ctx, conv)
	if err != nil {
		errMessage := fmt.Sprintf("%s\n```%s```", OnErrorMessage, err.Error())
		_ = botMessage.UpdateMessage(errMessage, false)
		return fmt.Errorf("failed to create chat completion stream: %v", err)
	}

	err = c.updateMessageWithChatStream(stream, botMessage)
	if err != nil {
		return fmt.Errorf("failed to update message with chat stream: %v", err)
	}

	return nil
}

func (c chat) updateMessageWithChatStream(stream *openai.ChatCompletionStream, message slackapi.BotMessage) error {
	nextUpdate := time.Now().Add(UpdateInterval)
	data := ""
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else if errors.Is(err, context.Canceled) {
				break
			} else {
				return fmt.Errorf("error on stream recv: %v", err)
			}
		}

		data += resp.Choices[0].Delta.Content
		if time.Now().After(nextUpdate) {
			err = message.UpdateMessage(data+UpdatingMessage, true)
			if err != nil {
				return fmt.Errorf("failed to update message: %v", err)
			}
			nextUpdate = time.Now().Add(UpdateInterval)
		}
	}

	err := message.UpdateMessage(data, false)
	if err != nil {
		return fmt.Errorf("failed to update message: %v", err)
	}

	return nil
}

func ProvideChat(
	gpt gpt.Client,
	config config.Config,
	logger logger.Logger,
	api slackapi.SlackAPI,
	crepo repository.ContextCancelRepository,
) Chat {
	return &chat{
		slack:  api,
		gpt:    gpt,
		config: config,
		logger: logger,
		crepo:  crepo,
	}
}
