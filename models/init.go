package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
)

func init() {
	// 注册所有的表，以便orm框架可以访问
	orm.RegisterModel(
		new(Category),
		new(Book),
		new(Document),
		new(Attachment),
		new(DocumentStore),
		new(BookCategory),
		new(Member),
		new(Collection),
		new(Relationship),
		new(Fans),
		new(Comments),
		new(Score),
	)
}

// 定义几个表名，通过函数来返回
// 将表名写进函数的原因：是平时开发的原则，开发过程中如果需要改表名，可以在一个地方统一地改
func TNCategory() string {
	return "md_category"
}

func TNBook() string {
	return "md_book"
}
func TNBookCategory() string {
	return "md_book_category"
}

func TNMembers() string {
	return "md_members"
}

func TNRelationship() string {
	return "md_relationship"
}

func TNDocuments() string {
	return "md_documents"
}

func TNComments() string {
	return "md_comments"
}

func TNScore() string {
	return "md_score"
}

func TNAttachment() string {
	return "md_attachment"
}

func TNDocumentStore() string {
	return "md_document_store"
}

/*
* Tool Funcs
* */
//设置增减
//@param            table           需要处理的数据表
//@param            field           字段
//@param            condition       条件
//@param            incre           是否是增长值，true则增加，false则减少
//@param            step            增或减的步长
func IncOrDec(table string, field string, condition string, incre bool, step ...int) (err error) {
	mark := "-"
	if incre {
		mark = "+"
	}
	s := 1
	if len(step) > 0 {
		s = step[0]
	}
	// update md_book set vcnt=vcnt+1 where book_id=doc.BookId 图书阅读人次+1
	// update md_book set vcnt=vcnt+1 where document_id=doc.DocumentId 章节阅读人次+1
	sql := fmt.Sprintf("update %v set %v=%v%v%v where %v", table, field, field, mark, s, condition)
	_, err = orm.NewOrm().Raw(sql).Exec()
	return
}
