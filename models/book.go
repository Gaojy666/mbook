package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
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

// Select 根据查询的字段和值来查询相应的数据，后面可以指定查询的字段结果
func (m *Book) Select(field string, value interface{}, cols ...string) (book *Book, err error) {
	if len(cols) == 0 {
		err = orm.NewOrm().QueryTable(m.TableName()).Filter(field, value).One(m)
	} else {
		err = orm.NewOrm().QueryTable(m.TableName()).Filter(field, value).One(m, cols...)
	}
	return m, err
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
