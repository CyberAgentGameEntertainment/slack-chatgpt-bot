package conversation

type (
	// Message - 会話に含まれる一つのメッセージ
	Message interface {
		Role() string
		Content() string
		SetContent(content string)
		UserName() string
		TimeStamp() string
	}

	message struct {
		role      string
		content   string
		username  string
		timeStamp string
	}
)

func (m *message) SetContent(content string) {
	m.content = content
}

func (m message) Role() string {
	return m.role
}

func (m message) Content() string {
	return m.content
}

func (m message) UserName() string {
	return m.username
}

func (m message) TimeStamp() string {
	return m.timeStamp
}

func NewMessage(role string, content string, username string, timeStamp string) Message {
	return &message{
		role:      role,
		content:   content,
		username:  username,
		timeStamp: timeStamp,
	}
}
