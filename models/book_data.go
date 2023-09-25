package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"time"
	"ziyoubiancheng/mbook/common"
)

// 拼接返回到接口的图书信息
// 在Book结构体的基础上，加入了用户与图书之间的关系的信息,少了ReleaseTime，多了7个字段
type BookData struct {
	BookId         int       `json:"book_id"`
	BookName       string    `json:"book_name"` //名称
	Identify       string    `json:"identify"`  //唯一标识
	OrderIndex     int       `json:"order_index"`
	Description    string    `json:"description"`     //图书描述
	PrivatelyOwned int       `json:"privately_owned"` //是否私有: 0 公开 ; 1 私有
	PrivateToken   string    `json:"private_token"`   //私有图书访问Token
	DocCount       int       `json:"doc_count"`       //文档数量
	CommentCount   int       `json:"comment_count"`
	CreateTime     time.Time `json:"create_time"`     //创建时间
	CreateName     string    `json:"create_name"`     // 对应member.account
	ModifyTime     time.Time `json:"modify_time"`     //修改时间
	Cover          string    `json:"cover"`           //封面地址
	MemberId       int       `json:"member_id"`       //用户Id
	Username       int       `json:"user_name"`       //
	Editor         string    `json:"editor"`          //编辑器类型: "markdown"
	RelationshipId int       `json:"relationship_id"` //
	RoleId         int       `json:"role_id"`         //
	RoleName       string    `json:"role_name"`       //
	Status         int       //状态:0 正常 ; 1 已删除
	Vcnt           int       `json:"vcnt"`             //阅读次数
	Collection     int       `json:"star"`             //收藏次数
	Score          int       `json:"score"`            //评分
	CntComment     int       `json:"cnt_comment"`      //评论人数
	CntScore       int       `json:"cnt_score"`        //评分人数
	ScoreFloat     string    `json:"score_float"`      // book.Score保留一位小数
	LastModifyText string    `json:"last_modify_text"` //
	Author         string    `json:"author"`           //来源
	AuthorURL      string    `json:"author_url"`       //来源链接
}

func NewBookData() *BookData {
	return &BookData{}
}

// SelectByIdentify 根据标识符和成员ID查询书籍数据信息
func (m *BookData) SelectByIdentify(identify string, memberId int) (result *BookData, err error) {
	if identify == "" || memberId <= 0 {
		return result, errors.New("Invalid parameter")
	}

	book := NewBook()
	o := orm.NewOrm()
	// 根据标识符查询书籍信息
	err = o.QueryTable(TNBook()).Filter("identify", identify).One(book)
	if err != nil {
		return
	}

	//查看权限
	relationship := NewRelationship()

	// 根据书籍ID和角色ID查询关系表信息，找到图书创始人的memberId
	// 如果找不到，则无论是谁，都没有权限获得图书信息
	err = o.QueryTable(TNRelationship()).Filter("book_id", book.BookId).Filter("role_id", 0).One(relationship)
	if err != nil {
		return result, errors.New("Permission denied")
	}

	// 根据memberId获得member的全部信息
	member, err := NewMember().Find(relationship.MemberId)
	if err != nil {
		return result, err
	}

	// 根据书籍ID和传进来的memberId来查询两者的关系
	err = o.QueryTable(TNRelationship()).Filter("book_id", book.BookId).Filter("member_id", memberId).One(relationship)
	if err != nil {
		return
	}

	// 将查询到的信息转换为 BookData 对象
	result = book.ToBookData()
	result.CreateName = member.Account
	result.MemberId = relationship.MemberId
	result.RoleId = relationship.RoleId
	result.RoleName = common.BookRole(result.RoleId)
	result.RelationshipId = relationship.RelationshipId

	document := NewDocument()
	// 根据书籍ID查询章节信息，并按修改时间排序
	// document并没有返回，返回err
	err = o.QueryTable(TNDocuments()).Filter("book_id", book.BookId).OrderBy("modify_time").One(document)
	return
}
