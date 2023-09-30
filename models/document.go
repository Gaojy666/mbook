package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"strings"
	"time"
	"ziyoubiancheng/mbook/utils"
)

// 图书章节内容
type Document struct {
	// orm:"column"表示会映射到数据库中的哪一列
	DocumentId   int           `orm:"pk;auto;column(document_id)" json:"doc_id"`                                    // 章节的唯一标识符
	DocumentName string        `orm:"column(document_name);size(500)" json:"doc_name"`                              // 章节名称
	Identify     string        `orm:"column(identify);size(100);index;null;default(null)" json:"identify"`          //唯一标识
	BookId       int           `orm:"column(book_id);type(int)" json:"book_id"`                                     //所属图书的唯一标识符
	ParentId     int           `orm:"column(parent_id);type(int);default(0)" json:"parent_id"`                      //父章节的唯一标识符。默认值为0，表示没有父章节。
	OrderSort    int           `orm:"column(order_sort);default(0);type(int)" json:"order_sort"`                    //章节的排序顺序
	Release      string        `orm:"column(release);type(text);null" json:"release"`                               // 章节的内容,发布后的HTML
	CreateTime   time.Time     `orm:"column(create_time);type(datetime);auto_now_add" json:"create_time"`           //章节创建的时间
	MemberId     int           `orm:"column(member_id);type(int)" json:"member_id"`                                 //创建章节的成员的唯一标识符
	ModifyTime   time.Time     `orm:"column(modify_time);type(datetime);default(null);auto_now" json:"modify_time"` //章节的修改时间
	ModifyAt     int           `orm:"column(modify_at);type(int)" json:"-"`                                         //修改操作的次数
	Version      int64         `orm:"type(bigint);column(version)" json:"version"`                                  //章节的版本号
	AttachList   []*Attachment `orm:"-" json:"attach"`                                                              //与章节关联的附件列表,数据表中没有该字段
	Vcnt         int           `orm:"column(vcnt);default(0)" json:"vcnt"`                                          //章节的访问计数，默认值为0
	Markdown     string        `orm:"-" json:"markdown"`                                                            //存储章节的 Markdown 格式内容。数据表中没有该字段
}

// 多字段唯一键
// 定义一组唯一组合索引
// beego框架会自动检查是否有TableUnique函数和TableIndex函数
// 自动为我们添加索引
func (m *Document) TableUnique() [][]string {
	return [][]string{
		[]string{"BookId", "Identify"},
	}
}

// 多字段索引
// 定义一组普通组合索引
func (m *Document) TableIndex() [][]string {
	return [][]string{
		[]string{"BookId", "ParentId", "OrderSort"},
	}
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
	if err == orm.ErrNoRows {
		return m, errors.New("数据不存在")
	}

	return m, nil
}

// 根据指定字段查询一条章节
func (m *Document) SelectByIdentify(BookId, Identify interface{}) (*Document, error) {
	err := GetOrm("r").
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

// 插入和更新文档
func (m *Document) InsertOrUpdate(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	id = int64(m.DocumentId)
	m.ModifyTime = time.Now()
	//去除 m.DocumentName 字段的前后空格。
	m.DocumentName = strings.TrimSpace(m.DocumentName)
	if m.DocumentId > 0 { //章节id存在，则更新
		_, err = o.Update(m, cols...)
		return
	}

	// 章节id不存在,说明是插入新的章节
	var selectDocument Document
	//  再根据identify和book_id去查询document_id
	o.QueryTable(TNDocuments()).Filter("identify", m.Identify).Filter("book_id", m.BookId).One(&selectDocument, "document_id")
	if selectDocument.DocumentId == 0 {
		m.CreateTime = time.Now()
		id, err = o.Insert(m)
		NewBook().RefreshDocumentCount(m.BookId)
	} else { //identify存在，则执行更新
		_, err = o.Update(m)
		id = int64(selectDocument.DocumentId)
	}
	return
}

// 发布文档内容
func (m *Document) ReleaseContent(bookId int, baseUrl string) {
	// 该函数实现功能的流程:
	// 1.上锁,标记为正在发布
	// 2.拿到该书的最多5000个章节的内容及其中每一个对应的附件列表
	// 3.遍历附件列表,取得附件地址,返回给前端html字符串形式,让前端进行拉取
	// 4.解锁,删除正在发布的标记

	// 将 bookId 标记为正在发布，以防止多处重复发布。
	utils.BooksRelease.Set(bookId)
	// 函数执行完毕后自动将 bookId 从正在发布的标记集合中删除。
	defer utils.BooksRelease.Delete(bookId)

	o := orm.NewOrm()
	var book Book
	querySeter := o.QueryTable(TNBook()).Filter("book_id", bookId)
	querySeter.One(&book)

	//重新发布
	var documents []*Document
	//过滤出指定 bookId 的最多 5000 条文档对象，并将结果存储在变量 documents 中
	_, err := o.QueryTable(TNDocuments()).Filter("book_id", bookId).Limit(5000).All(&documents, "document_id")
	if err != nil {
		return
	}

	documentStore := new(DocumentStore)
	for _, doc := range documents {
		//从 documentStore 中获取文档内容并去除首尾的空白字符
		content := strings.TrimSpace(documentStore.SelectField(doc.DocumentId, "content"))
		doc.Release = content
		//获取与当前文档关联的附件列表
		attachList, err := NewAttachment().SelectByDocumentId(doc.DocumentId)
		// 如果成功获取到附件列表且列表长度大于 0,生成一个 HTML 字符串，包含附件链接
		if err == nil && len(attachList) > 0 {
			content := bytes.NewBufferString("<div class=\"attach-list\"><strong>附件</strong><ul>")
			for _, attach := range attachList {
				li := fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\" title=\"%s\">%s</a></li>", attach.HttpPath, attach.Name, attach.Name)
				content.WriteString(li)
			}
			content.WriteString("</ul></div>")
			doc.Release += content.String()
		}
		o.Update(doc, "release")
	}

	//更新时间戳
	if _, err = querySeter.Update(orm.Params{
		"release_time": time.Now(),
	}); err != nil {
		logs.Error(err.Error())
	}
}

// 删除文档及其子文档
func (m *Document) Delete(docId int) error {

	o := orm.NewOrm()
	modelStore := new(DocumentStore)

	if doc, err := m.SelectByDocId(docId); err == nil {
		o.Delete(doc)
		modelStore.Delete(docId)
	}

	var docs []*Document

	_, err := o.QueryTable(m.TableName()).Filter("parent_id", docId).All(&docs)
	if err != nil {
		return err
	}

	for _, item := range docs {
		docId := item.DocumentId
		o.QueryTable(m.TableName()).Filter("document_id", docId).Delete()
		//删除document_store表对应的文档
		modelStore.Delete(docId)
		m.Delete(docId)
	}
	return nil
}
