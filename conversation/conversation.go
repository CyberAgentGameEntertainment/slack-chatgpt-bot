package conversation

import (
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
)

type (
	// Conversation - GPTに送信する一連の会話
	Conversation interface {
		// SystemMessage - システムメッセージを追加します
		SystemMessage(content string)

		// Messages - 会話のメッセージを取得します
		Messages() []Message

		// Tokens - 現在の会話のトークン数を取得します
		Tokens(model string) int

		// ToChatCompletionMessage - 会話をChatCompletionMessageに変換します
		ToChatCompletionMessage() []openai.ChatCompletionMessage

		// RemoveMessageAfterTimestamp - 指定したタイムスタンプ以降のメッセージを削除します
		RemoveMessageAfterTimestamp(timestamp string)

		// TrimMessagesToSaveToken - トークン数を指定した数になるまでメッセージを削除します
		TrimMessagesToSaveToken(model string, maxTokens int)
	}

	conv struct {
		system   string
		messages []Message
	}
)

func (c *conv) TrimMessagesToSaveToken(model string, maxTokens int) {
	for i := 0; i < len(c.messages); i++ {
		if c.Tokens(model) > maxTokens {
			c.messages[i].SetContent("deleted to save token")
		} else {
			break
		}
	}
}

func (c *conv) RemoveMessageByTimestamp(timestamp string) (Message, bool) {
	for i, m := range c.messages {
		if m.TimeStamp() == timestamp {
			c.messages = append(c.messages[:i], c.messages[i+1:]...)
			return m, true
		}
	}
	return nil, false
}

func (c *conv) RemoveMessageAfterTimestamp(timestamp string) {
	var messages []Message
	for _, m := range c.messages {
		if m.TimeStamp() == timestamp {
			break
		}
		messages = append(messages, m)
	}

	c.messages = messages
}

func (c *conv) Tokens(model string) int {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0
	}

	// NOTE: this number may not be accurate because it varies from model to model.
	tokensPerMessage := 3
	tokensPerName := 1

	var numTokens int
	if c.system != "" {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(c.system, nil, nil))
		numTokens += len(tkm.Encode(openai.ChatMessageRoleSystem, nil, nil))
	}

	for _, message := range c.messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content(), nil, nil))
		numTokens += len(tkm.Encode(message.Role(), nil, nil))
		numTokens += len(tkm.Encode(message.UserName(), nil, nil))
		if message.UserName() != "" {
			numTokens += tokensPerName
		}
	}

	numTokens += 3
	return numTokens
}

func (c *conv) Messages() []Message {
	return c.messages
}

func (c *conv) SystemMessage(content string) {
	c.system = content
}

func (c *conv) ToChatCompletionMessage() []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage

	if c.system != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: c.system,
		})
	}

	for _, m := range c.messages {
		// NEVER OUTPUT LOGS TO PROTECT PRIVACY
		// fmt.Printf("MESSAGE: %s\n", m.Content())

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    m.Role(),
			Content: m.Content(),
			Name:    m.UserName(),
		})
	}
	return messages
}

func NewConversation(messages []Message) Conversation {
	return &conv{
		messages: messages,
	}
}

func NewConversationFromSlackMessages(message []slack.Message, botUserID string) Conversation {
	var messages []Message
	for _, m := range message {
		if m.User == botUserID {
			if len(m.Blocks.BlockSet) >= 1 {
				block, ok := m.Blocks.BlockSet[0].(*slack.ActionBlock)
				if ok && block.Type == slack.MBTAction {
					continue // this is controller
				}
			}

			messages = append(messages, NewMessage(openai.ChatMessageRoleAssistant, m.Text, m.Username, m.Timestamp))
		} else {
			text := m.Text + " (UserID: <@" + m.User + ">)"
			messages = append(messages, NewMessage("user", text, m.Username, m.Timestamp))
		}
	}
	return NewConversation(messages)
}
