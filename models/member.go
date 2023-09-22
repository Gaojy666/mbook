package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/utils"
)

type Member struct {
	MemberId      int       `orm:"pk;auto" json:"member_id"`
	Account       string    `orm:"size(30);unique" json:"account"`
	Nickname      string    `orm:"size(30);unique" json:"nickname"`
	Password      string    ` json:"-"`
	Description   string    `orm:"size(640)" json:"description"`
	Email         string    `orm:"size(100);unique" json:"email"`
	Phone         string    `orm:"size(20);null;default(null)" json:"phone"`
	Avatar        string    `json:"avatar"`                // 头像
	Role          int       `orm:"default(1)" json:"role"` // 用户等级
	RoleName      string    `orm:"-" json:"role_name"`     // 等级名称
	Status        int       `orm:"default(0)" json:"status"`
	CreateTime    time.Time `orm:"type(datetime);auto_now_add" json:"create_time"`
	CreateAt      int       `json:"create_at"`
	LastLoginTime time.Time `orm:"type(datetime);null" json:"last_login_time"`
}

func (m *Member) TableName() string {
	return TNMembers()
}

func NewMember() *Member {
	return &Member{}
}

// 验证当前sessionId是否有效，即能否通过该sessionId查询到相应用户
func (m *Member) Find(id int) (*Member, error) {
	m.MemberId = id
	if err := orm.NewOrm().Read(m); err != nil {
		return m, err
	}
	m.RoleName = common.Role(m.Role)
	return m, nil
}

// 更新用户的相应字段
func (m *Member) Update(cols ...string) error {
	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := orm.NewOrm().Update(m, cols...); err != nil {
		return err
	}
	return nil
}

// 当cookie过期时需要登陆
func (m *Member) Login(account string, password string) (*Member, error) {
	member := &Member{}
	// 数据库中有对应的账户，并且为没有登陆状态，才是正常的。
	// 将结果存储到 member 中。
	err := orm.NewOrm().QueryTable(m.TableName()).Filter("account", account).Filter("status", 0).One(member)
	if err != nil {
		return member, errors.New("用户不存在")
	}

	// 验证password
	ok, err := utils.PasswordVerify(member.Password, password)
	// 如果password验证通过
	if ok && err == nil {
		m.RoleName = common.Role(m.Role)
		return member, nil
	}

	return member, errors.New("密码错误")
}
