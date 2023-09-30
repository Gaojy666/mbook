package models

import (
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
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
	BookId     int       `orm:"index"` //书id
	Content    string    //评论内容
	TimeCreate time.Time //评论时间
}

func (m *Comments) TableName() string {
	return TNComments()
}

// 添加一条评论
func (m *Comments) AddComments(uid, bookId int, content string) (err error) {
	// 将评论内容插入，并且评论数+1

	// 1.限制评论频率 （防止刷评论）
	var comment Comments
	second := 10
	// 从comments表中根据uid和time_create查询评论的用户id
	// order by id desc最后插入的评论肯定id最大
	sql := `select id from` + TNComments() + `where uid=? and time_create>? order by id desc`
	o := orm.NewOrm()
	// 用户上一次评论时间距离现在评论时间应该超过10s
	o.Raw(sql, uid, time.Now().Add(-time.Duration(second)*time.Second)).QueryRow(&comment)
	if comment.Id > 0 {
		// 如果存在这种记录(10s之内用户刚评论上一条)
		return errors.New(fmt.Sprintf("您距离上次发表评论的时间小于%v秒，请稍后再发", second))
	}
	// 2.插入评论数据
	sql = `insert into` + TNComments() + `(uid, book_id, content, time_create) values(?,?,?,?)`
	_, err = o.Raw(sql, uid, content, time.Now()).Exec()
	if err != nil {
		logs.Error(err.Error())
		err = errors.New("发表评论失败")
		return err
	}
	// 3.增加评论数（评论数+1）
	sql = `update ` + TNBook() + ` set cnt_comment=cnt_comment+1 where book_id=?`
	o.Raw(sql, bookId)

	return
}

// 评论内容
type BookCommentsResult struct {
	Uid        int       `json:"uid"`
	Score      int       `json:"score"`       //评分
	Avatar     string    `json:"avatar"`      //用户头像
	Nickname   string    `json:"nickname"`    //用户昵称
	Content    string    `json:"content"`     //评论内容
	TimeCreate time.Time `json:"time_create"` //评论时间
}

// 根据bookId评论内容，以及用户头像，昵称，评分, 评论时间
func (m *Comments) BookComments(page, size, bookId int) (comments []BookCommentsResult, err error) {
	//联合查询
	//从评论表中检索指定书籍ID的评论数据，并关联会员表和评分表，以获取评论内容、评分、会员信息等
	sql := `select c.content,s.score,c.uid,c.time_create,m.avatar,m.nickname
		from ` + TNComments() + ` c
		left join ` + TNMembers() + ` m on m.member_id=c.uid
		left join ` + TNScore() + ` s on s.uid=c.uid and s.book_id=c.book_id
		where c.book_id=? order by c.id desc limit %v offset %v`
	sql = fmt.Sprintf(sql, size, (page-1)*size)
	_, err = orm.NewOrm().Raw(sql, bookId).QueryRows(&comments)
	return

	// 指定数据库uar
	//o := orm.NewOrm()
	//
	//sql := `select book_id,uid,content,time_create from ` +
	//	TNComments() +
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
	//_, err = o.Raw(sql).QueryRows(&members)
	//if nil != err {
	//	return
	//}
	//// 接下来要将member中的nickname和avator放入comments中，以uid为桥梁
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

// 添加评分
// 一个用户只能评1次分
// score的值1-5, 然后对score*10, 50分表示5.0分
func (m *Score) AddScore(uid, bookId, score int) (err error) {
	// 查询评分是否已经存在
	o := orm.NewOrm()
	var scoreObj = Score{Uid: uid, BookId: bookId}
	// 先判断用户之前是否评过分
	o.Read(&scoreObj, "uid", "book_id")
	if scoreObj.Id > 0 {
		// 存在记录，说明之前评过分
		err = errors.New("您已经给当前图书打过分了")
		return
	}

	// 评分不存在，添加评分记录
	score = score * 10
	scoreObj.Score = score
	scoreObj.TimeCreate = time.Now()
	o.Insert(&scoreObj)
	if scoreObj.Id > 0 {
		// 评分添加成功，更新当前书籍的评分
		var book = Book{BookId: bookId}
		o.Read(&book, "book_id")
		book.CntScore = book.CntScore + 1
		// 更新平均分
		book.Score = (book.Score*(book.CntScore-1) + score) / book.CntScore
		// 更新评分次数和平均分
		_, err = o.Update(&book, "cnt_score", "score")
		if err != nil {
			logs.Error(err.Error())
			err = errors.New("评分失败，内部错误")
		}
	}
	return
}

// 查询该用户对文档的评分
func (m *Score) BookScoreByUid(uid, bookId interface{}) int {
	var score Score
	orm.NewOrm().QueryTable(TNScore()).Filter("uid", uid).Filter("book_id", bookId).One(&score, "score")
	return score.Score
}
