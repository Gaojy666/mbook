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

//func (m *Comments) TableName() string {
//	return TNComments()
//}

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
	sql := `select c.content,s.score,c.uid,c.time_create,m.avatar,m.nickname from ` + TNComments(bookId) + ` c left join ` + TNMembers() + ` m on m.member_id=c.uid left join ` + TNScore() + ` s on s.uid=c.uid and s.book_id=c.book_id where c.book_id=? order by c.id desc limit %v offset %v`
	sql = fmt.Sprintf(sql, size, (page-1)*size)
	_, err = GetOrm("w").Raw(sql, bookId).QueryRows(&comments)
	return

	//// 指定数据库uar
	//o := GetOrm("uar")
	//
	//sql := `select book_id,uid,content,time_create from ` +
	//	TNComments(bookId) +
	//	` where book_id=? limit %v offset %v`
	//sql = fmt.Sprintf(sql, size, (page-1)*size)
	//// 根据bookId去指定的评论数据库中查询用户uid,评论内容content,评论时间time_create
	//_, err = o.Raw(sql, bookId).QueryRows(&comments)
	//if nil != err {
	//	return
	//}
	//
	////头像昵称
	//uids := []string{}
	//for _, v := range comments {
	//	uids = append(uids, strconv.Itoa(v.Uid))
	//}
	//uidstr := strings.Join(uids, ",")
	//sql = `select member_id,avatar,nickname from md_members where member_id in(` + uidstr + `)`
	//members := []Member{}
	////提取评论者的用户ID，查询对应的用户信息（头像、昵称），并将结果存储到 members 切片中。
	//_, err = GetOrm("r").Raw(sql).QueryRows(&members)
	//if nil != err {
	//	return
	//}
	//memberMap := make(map[int]Member)
	////通过遍历 comments 切片和 members 切片，将用户的头像和昵称信息与评论信息关联起来。
	//for _, member := range members {
	//	memberMap[member.MemberId] = member
	//}
	//for k, v := range comments {
	//	comments[k].Avatar = memberMap[v.Uid].Avatar
	//	comments[k].Nickname = memberMap[v.Uid].Nickname
	//}
	//
	////评分
	////查询图书评论者的评分信息，并将结果存储到 scores 切片中。
	//sql = `select uid,score from md_score where book_id=? and uid in(` + uidstr + `)`
	//scores := []Score{}
	//_, err = o.Raw(sql, bookId).QueryRows(&scores)
	//if nil != err {
	//	return
	//}
	//scoreMap := make(map[int]Score)
	//for _, score := range scores {
	//	scoreMap[score.Uid] = score
	//}
	//for k, v := range comments {
	//	comments[k].Score = scoreMap[v.Uid].Score
	//}
	//
	//return
}
