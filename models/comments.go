package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

/*
*
*	评论
*
 */

// 评论表
type Comments struct {
	Id         int
	Uid        int       `orm:"index"` //用户id
	BookId     int       `orm:"index"` //文档项目id
	Content    string    //评论内容
	TimeCreate time.Time //评论时间
}

func (m *Comments) TableName() string {
	return TNComments()
}

/*
*
*	评分
*
 */

// 评分表
type Score struct {
	Id         int
	BookId     int
	Uid        int
	Score      int //评分
	TimeCreate time.Time
}

func (m *Score) TableName() string {
	return TNScore()
}

// 查询用户对文档的评分
func (m *Score) BookScoreByUid(uid, bookId interface{}) int {
	var score Score
	orm.NewOrm().QueryTable(TNScore()).Filter("uid", uid).Filter("book_id", bookId).One(&score, "score")
	return score.Score
}

// 评论内容
type BookCommentsResult struct {
	Uid        int       `json:"uid"`
	Score      int       `json:"score"`
	Avatar     string    `json:"avatar"`
	Nickname   string    `json:"nickname"`
	Content    string    `json:"content"`
	TimeCreate time.Time `json:"time_create"` //评论时间
}

// 评论内容
func (m *Comments) BookComments(page, size, bookId int) (comments []BookCommentsResult, err error) {
	sql := `select c.content, s.score, c.uid, c.time_create, m.avatar, m.nickname from ` +
		TNComments() + ` c left join ` + TNMembers() +
		`m on m.member_id=c.uid left join ` + TNScore() +
		`s on s.uid=c.uid and s.book_id=c_book_id where c.book_id=? order by c.id desc limit %v offset %v`
	// 按照页数查询数据
	sql = fmt.Sprintf(sql, size, (page-1)*size)
	// Raw方法执行原生SQL查询，将结果放入comments中
	_, err = orm.NewOrm().Raw(sql, bookId).QueryRows(&comments)
	return
}
