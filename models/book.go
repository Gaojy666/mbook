package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"strconv"
	"strings"
	"time"
	"ziyoubiancheng/mbook/utils"
)

type Book struct {
	BookId      int    `orm:"pk;auto" json:"book_id"` // 主键，自增
	BookName    string `orm:"size(500)" json:"book_name"`
	Identify    string `orm:"size(100);unique" json:"identify"` // 图书的唯一标识
	OrderIndex  int    `orm:"default(0)" json:"order_index"`    // 分类页下面，一个分类的多本图书有排序
	Description string `orm:"size(1000)" json:"description"`    // 图书的描述

	Cover          string    `orm:"size(1000)" json:"cover"`             // 封面地址
	Editor         string    `orm:"size(50)" json:"editor"`              //编辑器类型: "markdown"
	Status         int       `orm:"default(0)" json:"status"`            //状态:0 正常 ; 1 已删除
	PrivatelyOwned int       `orm:"default(0)" json:"privately_owned"`   // 是否私有: 0 公开 ; 1 私有
	PrivateToken   string    `orm:"size(500);null" json:"private_token"` // 私有图书访问Token
	MemberId       int       `orm:"size(100)" json:"member_id"`
	CreateTime     time.Time `orm:"type(datetime);auto_now_add" json:"create_time"` //创建时间
	ModifyTime     time.Time `orm:"type(datetime);auto_now_add" json:"modify_time"`
	ReleaseTime    time.Time `orm:"type(datetime);" json:"release_time"` //发布时间
	DocCount       int       `json:"doc_count"`                          //章节数量
	CommentCount   int       `orm:"type(int)" json:"comment_count"`      //评论数量
	Vcnt           int       `orm:"default(0)" json:"vcnt"`              //阅读人次
	Collection     int       `orm:"column(star);default(0)" json:"star"` //收藏次数
	Score          int       `orm:"default(40)" json:"score"`            //评分
	CntScore       int       //评分人数
	CntComment     int       //评论人数
	Author         string    `orm:"size(50)"`                      //来源
	AuthorURL      string    `orm:"column(author_url);size(1000)"` //来源链接
}

func (m *Book) TableName() string {
	return TNBook()
}

func NewBook() *Book {
	return &Book{}
}

// HomeData()将对应分类下所有的图书信息以及数据的个数返回
// fileds是查询指定的字段
func (m *Book) HomeData(pageIndex, pageSize int, cid int, fileds ...string) (books []Book, totalCount int, err error) {
	// 构造两个查询 一个是查询Book切片，一个是查询totalCount总共有多少本书

	// 首先查询Book切片的有关信息
	// 如果没有指定字段，那么默认查询5个字段
	if len(fileds) == 0 {
		fileds = append(fileds, "book_id", "book_name", "identify", "cover", "order_index")
	}
	// fields是一个数组，将其转化为字符串
	fieldStr := "b." + strings.Join(fileds, ",b.")

	sqlFmt := "select %v from " + TNBook() + " b left join " + TNBookCategory() + " c on b.book_id=c.book_id where c.category_id = " + strconv.Itoa(cid)
	sql := fmt.Sprintf(sqlFmt, fieldStr)

	// 而后查询totalCount
	sqlCount := fmt.Sprintf(sqlFmt, "count(*) cnt")

	o := orm.NewOrm()
	// 这里为什么要定义切片？
	var params []orm.Params
	if _, err := o.Raw(sqlCount).Values(&params); err == nil {
		if len(params) > 0 {
			totalCount, _ = strconv.Atoi(params[0]["cnt"].(string))
		}
	}

	_, err = o.Raw(sql).QueryRows(&books)

	return

}

// Select 根据查询的字段和值来查询Book数据，后面可以指定查询的字段结果
func (m *Book) Select(field string, value interface{}, cols ...string) (book *Book, err error) {
	if len(cols) == 0 {
		err = orm.NewOrm().QueryTable(m.TableName()).Filter(field, value).One(m)
	} else {
		err = orm.NewOrm().QueryTable(m.TableName()).Filter(field, value).One(m, cols...)
	}
	return m, err
}

func (m *Book) SelectPage(pageIndex, pageSize, memberId int, PrivatelyOwned int) (books []*BookData, totalCount int, err error) {
	o := orm.NewOrm()
	sql1 := "select count(b.book_id) as total_count from " + TNBook() + " as b left join " +
		TNRelationship() + " as r on b.book_id=r.book_id and r.member_id = ? where r.relationship_id > 0  and b.privately_owned=" + strconv.Itoa(PrivatelyOwned)

	err = o.Raw(sql1, memberId).QueryRow(&totalCount)
	if err != nil {
		return
	}
	offset := (pageIndex - 1) * pageSize

	//从 TNBook() 获取图书表名，并将其作为别名 book。
	//使用左连接将 TNRelationship() 表和 book 表关联起来，条件是两个表的 book_id 列相等，并且 rel 表的 member_id 列等于传递的 memberId 参数。
	//使用左连接将 TNRelationship() 表再次与 book 表关联起来，条件是两个表的 book_id 列相等，并且 rel1 表的 role_id 列等于 0。
	//使用左连接将 TNMembers() 表与 rel1 表关联起来，条件是 rel1 表的 member_id 列等于 m 表的 member_id 列。
	//添加过滤条件 where rel.relationship_id > 0，%v 是一个占位符，用于在后面进行动态替换。
	//添加排序条件 order by book.book_id desc。
	//添加分页限制条件 limit offset, pageSize，其中 offset 和 pageSize 是根据传递的 pageIndex 和 pageSize 参数计算得到的。
	sql2 := "select book.*,rel.member_id,rel.role_id,m.account as create_name from " + TNBook() + " as book" +
		" left join " + TNRelationship() + " as rel on book.book_id=rel.book_id and rel.member_id = ?" +
		" left join " + TNRelationship() + " as rel1 on book.book_id=rel1.book_id  and rel1.role_id=0" +
		" left join " + TNMembers() + " as m on rel1.member_id=m.member_id " +
		" where rel.relationship_id > 0 %v order by book.book_id desc limit " + fmt.Sprintf("%d,%d", offset, pageSize)
	sql2 = fmt.Sprintf(sql2, " and book.privately_owned="+strconv.Itoa(PrivatelyOwned))

	_, err = o.Raw(sql2, memberId).QueryRows(&books)
	if err != nil {
		return
	}
	return
}

func (book *Book) ToBookData() (m *BookData) {
	m = &BookData{
		BookId:         book.BookId,
		BookName:       book.BookName,
		Identify:       book.Identify,
		OrderIndex:     book.OrderIndex,
		Description:    strings.Replace(book.Description, "\r\n", "<br/>", -1),
		PrivatelyOwned: book.PrivatelyOwned,
		PrivateToken:   book.PrivateToken,
		DocCount:       book.DocCount,
		CommentCount:   book.CommentCount,
		CreateTime:     book.CreateTime,
		// CreateName
		ModifyTime: book.ModifyTime,
		Cover:      book.Cover,
		MemberId:   book.MemberId,
		// Username
		Editor: book.Editor,
		// RelationshipId
		// RoleId
		// RoleName
		Status:     book.Status,
		Vcnt:       book.Vcnt,
		Collection: book.Collection,
		Score:      book.Score,
		CntComment: book.CntComment,
		CntScore:   book.CntScore,
		ScoreFloat: utils.ScoreFloat(book.Score), //将int型保留一位小数
		// LastModifyText
		Author:    book.Author,
		AuthorURL: book.AuthorURL,
	}

	if book.Editor == "" {
		m.Editor = "markdown" // 默认Markdown编辑器
	}
	return m
}

// 更新章节数量
func (m *Book) RefreshDocumentCount(bookId int) {
	o := orm.NewOrm()
	// 查询该书中的章节数
	docCount, err := o.QueryTable(TNDocuments()).Filter("book_id", bookId).Count()
	if err == nil {
		temp := NewBook()
		temp.BookId = bookId
		temp.DocCount = int(docCount)
		// 已经指定了BookId,可以直接更行doc_count
		o.Update(temp, "doc_count")
	} else {
		logs.Error(err)
	}
}

// 根据书名和简介来查询书
func (m *Book) SearchBook(wd string, page, size int) (books []Book, cnt int, err error) {
	//收藏越多的图书放到最前面
	//select book_id from md_book where book_name like "%bo%" or description like "bo" order by star desc;
	sqlFmt := `select %v from md_books where book_name like ? or description like ? order by star desc;`
	sql := fmt.Sprintf(sqlFmt, "book_id")
	sqlCount := fmt.Sprintf(sqlFmt, "count(book_id) cnt")

	wd = "%" + wd + "%"
	o := orm.NewOrm()
	var count struct{ Cnt int }
	err = o.Raw(sqlCount, wd, wd).QueryRow(&count)
	if count.Cnt > 0 {
		cnt = count.Cnt
		_, err = o.Raw(sql+" limit ? offset ?", wd, wd, size, (page-1)*size).QueryRows(books)
	}
	return
}

// Insert
func (m *Book) Insert() (err error) {
	if _, err = orm.NewOrm().Insert(m); err != nil {
		return
	}

	relationship := Relationship{BookId: m.BookId, MemberId: m.MemberId, RoleId: 0}
	if err = relationship.Insert(); err != nil {
		return err
	}

	document := Document{BookId: m.BookId, DocumentName: "空白文档", Identify: "blank", MemberId: m.MemberId}
	var id int64
	if id, err = document.InsertOrUpdate(); err == nil {
		documentstore := DocumentStore{DocumentId: int(id), Markdown: ""}
		err = documentstore.InsertOrUpdate()
	}
	return err
}

// Update
func (m *Book) Update(cols ...string) (err error) {
	bk := NewBook()
	bk.BookId = m.BookId
	o := orm.NewOrm()
	if err = o.Read(bk); err != nil {
		return err
	}
	_, err = o.Update(m, cols...)
	return err
}
