package slackapi

import (
	"fmt"
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/slack-go/slack"
	"strings"
)

const (
	CustomInstructionsMarker = "SlackBot:"
)

type (
	SlackAPI interface {
		GetBotUserId() (string, error)
		LoadConversationReplies(channelId string, timeStamp string) ([]slack.Message, error)
		CreateNewBotMessage(channelId string, timeStamp string, msg string) (BotMessage, error)
		TakeOverBotMessage(channelId string, botMessageTS string, controllerTS string) (BotMessage, error)
		LoadCustomInstructions(channelId string) (string, error)
	}

	slackAPI struct {
		client *slack.Client
		logger logger.Logger
	}
)

func (s slackAPI) CreateNewBotMessage(channelId string, timeStamp string, msg string) (BotMessage, error) {
	return NewBotMessage(s.client, channelId, timeStamp, msg)
}

func (s slackAPI) TakeOverBotMessage(channelId string, botMessageTS string, controllerTS string) (BotMessage, error) {
	return TakeOverBotMessage(s.client, channelId, botMessageTS, controllerTS), nil
}

func (s slackAPI) LoadConversationReplies(channelId string, timeStamp string) ([]slack.Message, error) {
	s.logger.Log(logger.VERB, "app mention event: load conversation replies")
	var messages []slack.Message

	var cursor string = ""
	for {
		resp, hasMore, nextCursor, err := s.client.GetConversationReplies(&slack.GetConversationRepliesParameters{
			ChannelID: channelId,
			Timestamp: timeStamp,
			Cursor:    cursor,
		})

		if err != nil {
			return []slack.Message{}, fmt.Errorf("failed to get conversation history: %v", err)
		}

		messages = append(messages, resp...)

		if !hasMore {
			break
		}

		cursor = nextCursor
	}

	return messages, nil
}

func (s slackAPI) GetBotUserId() (string, error) {
	authTestResponse, err := s.client.AuthTest()
	if err != nil {
		return "", err
	}

	return authTestResponse.UserID, nil
}

func (s slackAPI) LoadCustomInstructions(channelId string) (string, error) {
	resp, err := s.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID:         channelId,
		IncludeLocale:     false,
		IncludeNumMembers: false,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get conversation info: %v", err)
	}

	ci := s.parseCustomInstructions(resp.Purpose.Value)
	if ci == "" {
		ci = s.parseCustomInstructions(resp.Topic.Value)
	}

	return ci, nil
}

func (s slackAPI) parseCustomInstructions(input string) string {
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		if strings.Contains(line, CustomInstructionsMarker) {
			return strings.TrimSpace(strings.SplitN(line, CustomInstructionsMarker, 2)[1])
		}
	}

	return ""
}

func ProvideSlackAPI(config config.Config, log logger.Logger) SlackAPI {
	client := slack.New(config.SlackBotToken())
	return &slackAPI{
		client: client,
		logger: log,
	}
}
