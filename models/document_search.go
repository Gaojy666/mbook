package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ziyoubiancheng/mbook/utils/html2text"
)

// 文档搜索结果
type DocumentData struct {
	DocumentId   int       `json:"doc_id"`
	DocumentName string    `json:"doc_name"`
	Identify     string    `json:"identify"`
	Release      string    `json:"release"` // Release 发布后的Html格式内容.
	Vcnt         int       `json:"vcnt"`    //文档图书被浏览次数
	CreateTime   time.Time `json:"create_time"`
	BookId       int       `json:"book_id"`
	BookIdentify string    `json:"book_identify"`
	BookName     string    `json:"book_name"`
}

// 文档搜索
type DocumentSearch struct {
	DocumentId   int       `json:"doc_id"`
	BookId       int       `json:"book_id"`
	DocumentName string    `json:"doc_name"`
	Identify     string    `json:"identify"` // Identify 文档唯一标识
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	BookName     string    `json:"book_name"`
	BookIdentify string    `json:"book_identify"`
	ModifyTime   time.Time `json:"modify_time"`
	CreateTime   time.Time `json:"create_time"`
}

func NewDocumentSearch() *DocumentSearch {
	return &DocumentSearch{}
}

// 图书内搜索.
func (m *DocumentSearch) SearchDocument(keyword string, bookId int, page, size int) (docs []*DocumentSearch, cnt int, err error) {
	//select * 部分
	//定义一个字符串切片 fields，包含需要查询的字段名。
	fields := []string{"document_id", "document_name", "identify", "book_id"}

	//构造sql,sqlcount
	var sql, sqlCount string
	if bookId == 0 {
		//如果 bookId 为 0，则查询所有图书的文档。
		//构造的 SQL 查询语句包括了联接图书表和文档表，
		//并使用 like 条件对文档名称和发布信息进行模糊匹配搜索。
		sql = "select %v from " + TNDocuments() + " d left join " + TNBook() + " b on d.book_id=b.book_id where b.privately_owned=0 and (d.document_name like ? or d.`release` like ? )"
		// select count(d.document_id) cnt
		// from md_documents d left join md_book b on d.book_id=b.book_id
		// where b.privately_owned=0 and (d.document_name like ? or d.`release` like ?)
		sqlCount = fmt.Sprintf(sql, "count(d.document_id) cnt")
		// select d.document_id, d.document_name, d.identify, d.book_id
		// order by d.vcnt desc
		// from md_documents d left join md_book b on d.book_id=b.book_id
		// where b.privately_owned=0 and (d.document_name like ? or d.`release` like ?)
		sql = fmt.Sprintf(sql, "d."+strings.Join(fields, ",d.")) + " order by d.vcnt desc"
	} else {
		//如果 bookId 不为 0，则仅查询指定图书ID的文档。
		//构造的 SQL 查询语句只涉及文档表，
		//并使用 like 条件对文档名称和发布信息进行模糊匹配搜索。
		sql = "select %v from " + TNDocuments() + " where book_id = " + strconv.Itoa(bookId) + " and (document_name like ? or `release` like ?) "
		// select count(document_id) cnt
		// from md_documents
		// where book_id = bookId  and (document_name like ? or `release` like ?)
		sqlCount = fmt.Sprintf(sql, "count(document_id) cnt")
		// select document_id, document_name, identify, book_id
		// order by vcnt desc
		// from md_documents
		// where book_id = bookId  and (document_name like ? or `release` like ?)
		sql = fmt.Sprintf(sql, strings.Join(fields, ",")) + " order by vcnt desc"
	}

	//用于存储查询结果的文档数量
	var count struct {
		Cnt int
	}
	like := "%" + keyword + "%"

	o := GetOrm("r")
	o.Raw(sqlCount, like, like).QueryRow(&count)
	cnt = count.Cnt
	limit := fmt.Sprintf(" limit %v offset %v", size, (page-1)*size)
	if cnt > 0 {
		_, err = o.Raw(sql+limit, like, like).QueryRows(&docs)
	}
	return
}

// 返回文档
func (m *DocumentSearch) GetDocsById(id []int, withoutCont ...bool) (docs []DocumentData, err error) {
	if len(id) == 0 {
		return
	}

	var idArr []string
	for _, i := range id {
		idArr = append(idArr, fmt.Sprint(i))
	}

	fields := []string{
		"d.document_id", "d.document_name", "d.identify", "d.vcnt", "d.create_time", "b.book_id",
	}

	// 不返回内容
	if len(withoutCont) == 0 || !withoutCont[0] {
		fields = append(fields, "b.identify book_identify", "d.release", "b.book_name")
	}

	sqlFmt := "select " + strings.Join(fields, ",") + " from " + TNDocuments() + " d left join md_books b on d.book_id=b.book_id where d.document_id in(%v)"
	sql := fmt.Sprintf(sqlFmt, strings.Join(idArr, ","))

	var rows []DocumentData
	var cnt int64

	cnt, err = GetOrm("r").Raw(sql).QueryRows(&rows)
	if cnt > 0 {
		docMap := make(map[int]DocumentData)
		for _, row := range rows {
			docMap[row.DocumentId] = row
		}

		for _, i := range id {
			if doc, ok := docMap[i]; ok {
				doc.Release = html2text.Html2Text(doc.Release)
				docs = append(docs, doc)
			}
		}
	}

	return
}
