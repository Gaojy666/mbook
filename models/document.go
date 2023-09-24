package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

// 图书章节内容
type Document struct {
	DocumentId   int           `orm:"pk;auto;column(document_id)" json:"doc_id"`
	DocumentName string        `orm:"column(document_name);size(500)" json:"doc_name"`
	Identify     string        `orm:"column(identify);size(100);index;null;default(null)" json:"identify"`
	BookId       int           `orm:"column(book_id);type(int)" json:"book_id"`
	ParentId     int           `orm:"column(parent_id);type(int);default(0)" json:"parent_id"`
	OrderSort    int           `orm:"column(order_sort);default(0);type(int)" json:"order_sort"`
	Release      string        `orm:"column(release);type(text);null" json:"release"`
	CreateTime   time.Time     `orm:"column(create_time);type(datetime);auto_now_add" json:"create_time"`
	MemberId     int           `orm:"column(member_id);type(int)" json:"member_id"`
	ModifyTime   time.Time     `orm:"column(modify_time);type(datetime);default(null);auto_now" json:"modify_time"`
	ModifyAt     int           `orm:"column(modify_at);type(int)" json:"-"`
	Version      int64         `orm:"type(bigint);column(version)" json:"version"`
	AttachList   []*Attachment `orm:"-" json:"attach"`
	Vcnt         int           `orm:"column(vcnt);default(0)" json:"vcnt"`
	Markdown     string        `orm:"-" json:"markdown"`
}

func (m *Document) TableName() string {
	return TNDocuments()
}

func NewDocument() *Document {
	return &Document{
		Version: time.Now().Unix(),
	}
}

// 根据指定字段查询一条章节
func (m *Document) SelectByIdentify(BookId, Identify interface{}) (*Document, error) {
	err := orm.NewOrm().
		QueryTable(m.TableName()).
		Filter("BookId", BookId).
		Filter("Identify", Identify).One(m)
	return m, err
}

// 获取图书目录
func (m *Document) GetMenuTop(bookId int) (docs []*Document, err error) {
	var docsAll []*Document
	o := orm.NewOrm()

	// 指定要查询的字段
	cols := []string{"document_id", "document_name", "member_id", "parent_id", "book_id", "identify"}

	fmt.Println("--------------start")

	_, err = o.QueryTable(m.TableName()).
		Filter("book_id", bookId).
		Filter("parent_id", 0).               //根据 parent_id 进行过滤，只获取一级目录
		OrderBy("order_sort", "document_id"). // 按照document_id进行排序
		Limit(5000).                          // 限制最大查询数量为 5000
		All(&docsAll, cols...)                // 执行查询并将结果存储到 docsAll 中

	fmt.Println("--------------end")

	for _, doc := range docsAll {
		docs = append(docs, doc)
	}
	return
}
