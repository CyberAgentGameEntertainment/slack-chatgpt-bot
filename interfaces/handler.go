package interfaces

import (
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/SGE-AI/sge-bot/usecase"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	EventHandler interface {
		HandleAppMentionEvent(innerEvent slackevents.AppMentionEvent) error
		HandleMessageEvent(innerEvent slackevents.MessageEvent) error
		HandleBlockActionsEvent(innerEvent slack.InteractionCallback) error
	}

	eventHandler struct {
		config config.Config
		logger logger.Logger
		slack  slack.Client
		chat   usecase.Chat
		stat   usecase.Statistics
	}
)

// HandleMessageEvent - メッセージを受け取り、会話を開始します (DM向け)
func (e eventHandler) HandleMessageEvent(event slackevents.MessageEvent) error {
	if event.ChannelType != "im" {
		return nil
	}

	if event.User == "" {
		return nil
	}

	if event.BotID != "" {
		return nil
	}

	ts := event.ThreadTimeStamp
	if ts == "" {
		ts = event.TimeStamp
	}

	e.logger.Log(logger.INFO, "start normal conversation userid by message event: %s", event.User)
	go e.stat.UsedBy(event.User)
	return e.chat.StartNormalConversation(event.Channel, ts)
}

// HandleAppMentionEvent - メンションを受け取り、会話を開始します (チャンネル向け)
func (e eventHandler) HandleAppMentionEvent(event slackevents.AppMentionEvent) error {
	ts := event.ThreadTimeStamp
	if ts == "" {
		ts = event.TimeStamp
	}

	e.logger.Log(logger.INFO, "start normal conversation userid by app mention event: %s", event.User)

	go e.stat.UsedBy(event.User)
	return e.chat.StartNormalConversation(event.Channel, ts)
}

// HandleBlockActionsEvent - ブロックアクションを受け取り、会話を再生成します
func (e eventHandler) HandleBlockActionsEvent(event slack.InteractionCallback) error {
	if event.Type != slack.InteractionTypeBlockActions {
		return nil
	}

	for _, action := range event.ActionCallback.BlockActions {
		if action.ActionID == "regenerate" {
			e.logger.Log(logger.INFO, "regenerate message userid: %s, timestamp: %s", event.User.ID, action.BlockID)
			return e.chat.RegenerateMessage(event.Channel.ID, action.BlockID, event.Container.ThreadTs, event.Container.MessageTs)
		} else if action.ActionID == "stop" {
			e.logger.Log(logger.INFO, "stop message userid: %s, timestamp: %s", event.User.ID, action.BlockID)
			return e.chat.StopGenerateMessage(event.Channel.ID, action.BlockID, event.Container.MessageTs)
		} else if action.ActionID == "delete" {
			e.logger.Log(logger.INFO, "delete message userid: %s, timestamp: %s", event.User.ID, action.BlockID)
			return e.chat.DeleteMessage(event.Channel.ID, action.BlockID, event.Container.MessageTs)
		} else {
			e.logger.Log(logger.INFO, "unknown action: %s", action.ActionID)
		}
	}

	return nil
}

func ProvideEventHandler(config config.Config, log logger.Logger, chat usecase.Chat, stat usecase.Statistics) EventHandler {
	return &eventHandler{
		config: config,
		logger: log,
		chat:   chat,
		stat:   stat,
	}
}
