package models

import "github.com/beego/beego/v2/client/orm"

// 文档编辑
type DocumentStore struct {
	DocumentId int    `orm:"pk;auto;column(document_id)"`
	Markdown   string `orm:"type(text);"` //markdown内容
	Content    string `orm:"type(text);"` //html内容
}

func (m *DocumentStore) TableName() string {
	return TNDocumentStore()
}

func (m *DocumentStore) SelectField(docId interface{}, field string) string {
	var ds = Document{}
	if field != "markdown" {
		field = "content"
	}
	orm.NewOrm().QueryTable(TNDocuments()).Filter("document_id", docId)
}
