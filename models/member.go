package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"regexp"
	"strings"
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

// 在这里获得与member结构体有关的表名,因此在本文件下
// 对该表进行查询时，不再需要指定表名
func (m *Member) TableName() string {
	return TNMembers()
}

func NewMember() *Member {
	return &Member{}
}

// 验证当前memberId是否有效，即能否通过该memberId查询到相应用户
func (m *Member) Find(id int) (*Member, error) {
	m.MemberId = id
	if err := GetOrm("w").Read(m); err != nil {
		return m, err
	}
	m.RoleName = common.Role(m.Role)
	return m, nil
}

// 添加注册的用户信息
func (m *Member) Add() error {
	if m.Email == "" {
		return errors.New("请填写邮箱")
	}
	if ok, err := regexp.MatchString(common.RegexpEmail, m.Email); !ok || err != nil {
		return errors.New("邮箱格式错误")
	}
	if l := strings.Count(m.Password, ""); l < 6 || l >= 20 {
		return errors.New("密码请输入6-20个字符")
	}

	// 只是创建了一个查询对象，尚未指定表
	cond := orm.NewCondition().Or("email", m.Email).Or("nickname", m.Nickname).Or("account", m.Account)
	var one Member
	o := GetOrm("w")

	if o.QueryTable(m.TableName()).SetCond(cond).One(&one, "member_id", "nickname", "account", "email"); one.MemberId > 0 {
		// 根据nickname, email, account来查询数据库，有至少一条存在
		// 下面分别进行判断
		if one.Nickname == m.Nickname {
			return errors.New("昵称已存在")
		}
		if one.Email == m.Email {
			return errors.New("邮箱已存在")
		}
		if one.Account == m.Account {
			return errors.New("用户已存在")
		}
	}

	// 对密码进行哈希加密
	hash, err := utils.PasswordHash(m.Password)

	if err != nil {
		return err
	}

	m.Password = hash
	// 将加密密码后的用户信息存储到数据库中
	_, err = o.Insert(m)
	if err != nil {
		logs.Error(err.Error())
		return err
	}

	m.RoleName = common.Role(m.Role)
	return nil
}

// 更新用户的相应字段
func (m *Member) Update(cols ...string) error {
	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := GetOrm("w").Update(m, cols...); err != nil {
		return err
	}
	return nil
}

// 当cookie过期时需要登陆
func (m *Member) Login(account string, password string) (*Member, error) {
	member := &Member{}
	// 数据库中有对应的等级，并且为没有登陆状态，才是正常的。
	// .one将查询结果存储到 member 中。
	err := GetOrm("w").QueryTable(m.TableName()).Filter("account", account).Filter("status", 0).One(member)
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

func (m *Member) IsAdministrator() bool {
	if m == nil || m.MemberId <= 0 {
		return false
	}
	return m.Role == 0 || m.Role == 1
}

// 获取用户名
func (m *Member) GetUsernameByUid(id interface{}) string {
	var user Member
	GetOrm("w").QueryTable(TNMembers()).Filter("member_id", id).One(&user, "account")
	return user.Account
}

// 获取昵称
func (m *Member) GetNicknameByUid(id interface{}) string {
	var user Member
	if err := GetOrm("w").QueryTable(TNMembers()).Filter("member_id", id).One(&user, "nickname"); err != nil {
		logs.Error(err.Error())
	}

	return user.Nickname
}

// 根据用户名获取用户信息
func (m *Member) GetByUsername(username string) (member Member, err error) {
	err = GetOrm("w").QueryTable(TNMembers()).Filter("account", username).One(&member)
	return
}
