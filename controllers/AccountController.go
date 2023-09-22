package controllers

import (
	"errors"
	"fmt"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
)

type AccountController struct {
	BaseController
}

func init() {

}

// 登录
func (c *AccountController) Login() {
	var remember CookieRemember

	// 验证cookie
	// GetSecureCookie 从已编码的浏览器 cookie 值返回已解码的 cookie 值。
	// 采用了sha256来作为加密算法，
	// 第一个参数secret是加密的密钥，第二个参数key是cookie的名字，返回cookie值
	if cookie, ok := c.GetSecureCookie(common.AppKey(), "login"); ok {
		// 如果cookie不能解码到remember结构体中，则有错误
		if err := utils.Decode(cookie, &remember); err == nil {
			if err = c.login(remember.MemberId); err == nil {
				return
			}
		}
	}
	c.TplName = "account/login.html"

	// 当前请求是否为Post请求
	if c.Ctx.Input.IsPost() {
		account := c.GetString("account")
		password := c.GetString("password")
		member, err := models.NewMember().Login(account, password)
		fmt.Println(err)
		if err != nil {
			c.JsonResult(1, "登录失败", nil)
		}
		member.LastLoginTime = time.Now()
		member.Update()
		v, err := utils.Encode(remember)
		if err != nil {
			c.SetSecureCookie(common.AppKey(), "login", v, 20*3600*365)
		}
		c.JsonResult(0, "ok")
	}
}

// 判断数据库是否有相应的sessionId
func (c *AccountController) login(memberId int) (err error) {
	member, err := models.NewMember().Find(memberId)
	if member.MemberId == 0 {
		errors.New("用户不存在")
	}
	// 如果没有数据
	if err != nil {
		return err
	}
	// 更新登陆时间
	member.LastLoginTime = time.Now()
	member.Update()
	c.SetMember(*member)

	var remember CookieRemember
	remember.MemberId = member.MemberId
	remember.Account = member.Account
	remember.Time = time.Now()
	v, err := utils.Encode(remember)
	if err != nil {
		c.SetSecureCookie(common.AppKey(), "login", v, 20*3600*365)
	}
	return err
}
