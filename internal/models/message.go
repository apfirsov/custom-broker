package models

type Message struct {
	Content []byte
}

func (m *Message) SetContent(Content []byte) {
	m.Content = Content
}

func (m *Message) GetContent() []byte {
	return m.Content
}

type ConsumerMessage struct {
	Message *Message
	AckChan chan interface{}
}

func (cm *ConsumerMessage) GetContent() []byte {
	return cm.Message.GetContent()
}

func (cm *ConsumerMessage) Ack() {
	select {
	case cm.AckChan <- nil:
	default:
	}
}
