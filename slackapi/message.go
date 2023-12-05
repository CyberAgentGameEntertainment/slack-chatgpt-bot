package slackapi

import (
	"fmt"
	"github.com/slack-go/slack"
)

type (
	BotMessage interface {
		UpdateMessage(message string, isUpdating bool) error

		Regenerate(initialMessage string)

		OutputTimeStamp() string

		DeleteMySelf() error
	}

	botMessage struct {
		webapi       *slack.Client
		channelID    string
		outputTS     string
		controllerTS string
	}
)

func (b botMessage) Regenerate(msg string) {
	go b.webapi.UpdateMessage(
		b.channelID,
		b.controllerTS,
		slack.MsgOptionBlocks(buildActionBlock(true, b.outputTS)),
	)

	go b.webapi.UpdateMessage(
		b.channelID,
		b.outputTS,
		slack.MsgOptionText(msg, false),
	)
}

func (b botMessage) DeleteMySelf() error {
	go b.webapi.DeleteMessage(b.channelID, b.outputTS)
	go b.webapi.DeleteMessage(b.channelID, b.controllerTS)
	return nil
}

func (b botMessage) OutputTimeStamp() string {
	return b.outputTS
}

func (b botMessage) UpdateMessage(message string, isUpdating bool) error {
	if !isUpdating {
		go b.webapi.UpdateMessage(
			b.channelID,
			b.controllerTS,
			slack.MsgOptionBlocks(buildActionBlock(false, b.outputTS)),
		)
	}

	_, _, _, err := b.webapi.UpdateMessage(b.channelID, b.outputTS, slack.MsgOptionText(message, false))

	// マークダウン形式に対応できるが、長い出力の場合Slackの仕様上See Moreを押さないと表示されない😕
	// この問題が解決できるまではプレーンテキストで対応する
	//_, _, _, err := b.webapi.UpdateMessage(
	//	b.channelID,
	//	b.outputTS,
	//	slack.MsgOptionBlocks(
	//		b.buildActionBlock(isUpdating),
	//		slack.NewSectionBlock(
	//			slack.NewTextBlockObject(
	//				slack.MarkdownType,
	//				message,
	//				false,
	//				false,
	//			),
	//			nil,
	//			nil,
	//		),
	//	),
	//)

	return err
}

func buildActionBlock(addStopButton bool, targetTimeStamp string) *slack.ActionBlock {
	var elements []slack.BlockElement
	if addStopButton {
		elements = append(elements, slack.NewButtonBlockElement(
			"stop",
			"stop",
			slack.NewTextBlockObject(
				slack.PlainTextType,
				":x: 停止",
				true,
				false,
			),
		))
	}
	elements = append(elements, slack.NewButtonBlockElement(
		"regenerate",
		"regenerate",
		slack.NewTextBlockObject(
			slack.PlainTextType,
			":recycle: 再生成",
			true,
			false,
		),
	))
	elements = append(elements, slack.NewButtonBlockElement(
		"delete",
		"delete",
		slack.NewTextBlockObject(
			slack.PlainTextType,
			":fire: 削除",
			true,
			false,
		),
	))

	return slack.NewActionBlock(targetTimeStamp, elements...)
}

func TakeOverBotMessage(webapi *slack.Client, channelID string, outputTS string, controllerTS string) BotMessage {
	return &botMessage{
		webapi:       webapi,
		channelID:    channelID,
		outputTS:     outputTS,
		controllerTS: controllerTS,
	}
}

func NewBotMessage(webapi *slack.Client, channelID string, threadTS string, initMessage string) (BotMessage, error) {
	_, respTimeStamp, err := webapi.PostMessage(
		channelID,
		slack.MsgOptionText(initMessage, false),
		slack.MsgOptionTS(threadTS),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to post message: %v", err)
	}

	// ActionBlockをBotMessageに追加することもできるが、See More問題があるため別のポストで行っている
	// See More問題を解決できたらこのコントローラーの仕組みは消すことができる
	_, controllerMessageTS, err := webapi.PostMessage(
		channelID,
		slack.MsgOptionTS(threadTS),
		slack.MsgOptionBlocks(buildActionBlock(true, respTimeStamp)),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to post message: %v", err)
	}

	return &botMessage{
		webapi:       webapi,
		channelID:    channelID,
		outputTS:     respTimeStamp,
		controllerTS: controllerMessageTS,
	}, nil
}
