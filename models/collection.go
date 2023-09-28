package models

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"strconv"
)

type CollectionData struct {
	BookId      int    `json:"book_id"`
	BookName    string `json:"book_name"`
	Identify    string `json:"identify"`
	Description string `json:"description"`
	DocCount    int    `json:"doc_count"`
	Cover       string `json:"cover"`
	MemberId    int    `json:"member_id"`
	Nickname    string `json:"user_name"`
	Vcnt        int    `json:"vcnt"`
	Collection  int    `json:"star"`
	Score       int    `json:"score"`
	CntComment  int    `json:"cnt_comment"`
	CntScore    int    `json:"cnt_score"`
	ScoreFloat  string `json:"score_float"`
	OrderIndex  int    `json:"order_index"`
}

type Collection struct {
	Id       int
	MemberId int `orm:"index"`
	BookId   int
}

func (m *Collection) TableName() string {
	return TNCollection()
}

// 收藏或取消收藏
// @param            uid         用户id
// @param            bid         书籍id
// @return           cancel      是否是取消收藏
func (m *Collection) Collection(uid, bid int) (cancel bool, err error) {
	var star = Collection{MemberId: uid, BookId: bid}
	o := orm.NewOrm()
	qs := o.QueryTable(TNCollection())
	o.Read(&star, "MemberId", "BookId")
	if star.Id > 0 { //取消收藏
		if _, err = qs.Filter("id", star.Id).Delete(); err == nil {
			IncOrDec(TNBook(), "star", fmt.Sprintf("book_id=%v and star>0", bid), false, 1)
		}
		cancel = true
	} else { //添加收藏
		cancel = false
		if _, err = o.Insert(&star); err == nil {
			//收藏计数+1
			IncOrDec(TNBook(), "star", "book_id="+strconv.Itoa(bid), true, 1)
		}
	}
	return
}
