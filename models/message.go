package models

// 简单的消息体
type Message struct {
	BookId  int
	BaseUrl string
}

func NewMessage(bookId int, baseUrl string) *Message {
	return &Message{
		BookId:  bookId,
		BaseUrl: baseUrl,
	}
}
