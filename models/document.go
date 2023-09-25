package models

import (
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

// 图书章节内容
type Document struct {
	// orm:"column"表示会映射到数据库中的哪一列
	DocumentId   int           `orm:"pk;auto;column(document_id)" json:"doc_id"`                                    // 章节的唯一标识符
	DocumentName string        `orm:"column(document_name);size(500)" json:"doc_name"`                              // 章节名称
	Identify     string        `orm:"column(identify);size(100);index;null;default(null)" json:"identify"`          //标识符
	BookId       int           `orm:"column(book_id);type(int)" json:"book_id"`                                     //所属图书的唯一标识符
	ParentId     int           `orm:"column(parent_id);type(int);default(0)" json:"parent_id"`                      //父章节的唯一标识符。默认值为0，表示没有父章节。
	OrderSort    int           `orm:"column(order_sort);default(0);type(int)" json:"order_sort"`                    //章节的排序顺序
	Release      string        `orm:"column(release);type(text);null" json:"release"`                               // 章节的内容
	CreateTime   time.Time     `orm:"column(create_time);type(datetime);auto_now_add" json:"create_time"`           //章节创建的时间
	MemberId     int           `orm:"column(member_id);type(int)" json:"member_id"`                                 //创建章节的成员的唯一标识符
	ModifyTime   time.Time     `orm:"column(modify_time);type(datetime);default(null);auto_now" json:"modify_time"` //章节的修改时间
	ModifyAt     int           `orm:"column(modify_at);type(int)" json:"-"`                                         //修改操作的次数
	Version      int64         `orm:"type(bigint);column(version)" json:"version"`                                  //章节的版本号
	AttachList   []*Attachment `orm:"-" json:"attach"`                                                              //与章节关联的附件列表
	Vcnt         int           `orm:"column(vcnt);default(0)" json:"vcnt"`                                          //章节的访问计数，默认值为0
	Markdown     string        `orm:"-" json:"markdown"`                                                            //存储章节的 Markdown 格式内容。
}

func (m *Document) TableName() string {
	return TNDocuments()
}

func NewDocument() *Document {
	return &Document{
		Version: time.Now().Unix(),
	}
}

// 根据章节Id查询指定章节
func (m *Document) SelectByDocId(id int) (doc *Document, err error) {
	if id <= 0 {
		return m, errors.New("invalid parameter")
	}

	o := orm.NewOrm()
	err = o.QueryTable(m.TableName()).Filter("document_id", id).One(m)
	if err == nil {
		return m, errors.New("数据不存在")
	}

	return m, nil
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
