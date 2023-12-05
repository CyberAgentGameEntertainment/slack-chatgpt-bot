package interfaces

import (
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type (
	SocketConnection interface {
		Run() error
	}

	socketConnection struct {
		webApi  *slack.Client
		handler EventHandler
		logger  logger.Logger
	}
)

// Run - ソケット接続を開始し、イベントをルーティングします
func (s socketConnection) Run() error {
	socketMode := socketmode.New(s.webApi)
	go func() {
		for envelope := range socketMode.Events {
			switch envelope.Type {
			case socketmode.EventTypeEventsAPI:
				s.logger.Log(logger.VERB, "events api event received")
				go s.handleEventTypeEventsAPI(socketMode, envelope)
			case socketmode.EventTypeInteractive:
				s.logger.Log(logger.VERB, "interactive event received")
				go s.handleInteractiveEvent(socketMode, envelope)
			case socketmode.EventTypeConnecting:
				s.logger.Log(logger.VERB, "connecting to slack...")
			case socketmode.EventTypeConnectionError:
				s.logger.Log(logger.ERROR, "connection error: %v", envelope.Data)
			case socketmode.EventTypeConnected:
				s.logger.Log(logger.INFO, "connected to slack! waiting for events...")
			case socketmode.EventTypeHello:
				s.logger.Log(logger.VERB, "hello slack!") // this is the first event received when connecting
			default:
				s.logger.Log(logger.VERB, "unexpected event type received: %s", envelope.Type)
			}
		}
	}()

	return socketMode.Run()
}

// handleEventTypeEventsAPI - EventTypeEventsAPIを処理します
func (s socketConnection) handleEventTypeEventsAPI(client *socketmode.Client, envelope socketmode.Event) {
	eventsAPIEvent, ok := envelope.Data.(slackevents.EventsAPIEvent)
	if !ok {
		s.logger.Log(logger.VERB, "unexpected event type received: %s", envelope.Type)
		return
	}

	client.Ack(*envelope.Request)

	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		s.logger.Log(logger.VERB, "app mention from user %s received", event.User)
		err := s.handler.HandleAppMentionEvent(*event)
		if err != nil {
			s.logger.Log(logger.ERROR, "failed to handle app mention event: %v", err)
		}
	case *slackevents.MessageEvent:
		s.logger.Log(logger.VERB, "message from user %s received", event.User)
		err := s.handler.HandleMessageEvent(*event)
		if err != nil {
			s.logger.Log(logger.ERROR, "failed to handle app mention event: %v", err)
		}
	default:
		s.logger.Log(logger.VERB, "unexpected event type received: %s", envelope.Type)
	}
}

// handleInteractiveEvent - EventTypeInteractiveを処理します
func (s socketConnection) handleInteractiveEvent(client *socketmode.Client, envelope socketmode.Event) {
	event, ok := envelope.Data.(slack.InteractionCallback)
	if !ok {
		s.logger.Log(logger.VERB, "unexpected event type received: %s", envelope.Type)
		return
	}
	client.Ack(*envelope.Request)

	switch event.Type {
	case slack.InteractionTypeBlockActions:
		s.logger.Log(logger.VERB, "block action from user %s received", event.User.ID)
		err := s.handler.HandleBlockActionsEvent(event)
		if err != nil {
			s.logger.Log(logger.ERROR, "failed to handle block action event: %v", err)
		}
	default:
		s.logger.Log(logger.VERB, "unexpected event type received: %s", envelope.Type)
	}
}

func ProvideSocketConnection(config config.Config, handler EventHandler, logger logger.Logger) SocketConnection {
	botToken := config.SlackBotToken()
	appLevelToken := config.SlackAppLevelToken()
	webApi := slack.New(botToken, slack.OptionAppLevelToken(appLevelToken))

	return &socketConnection{
		webApi:  webApi,
		handler: handler,
		logger:  logger,
	}
}
