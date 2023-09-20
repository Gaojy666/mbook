package models

type BookCategory struct {
	Id       int // 自增主键
	BookId   int // book id
	Category int // 分类 id
}

func (m *BookCategory) TableName() string {
	return TNBookCategory()
}
