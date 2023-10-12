package controllers

import (
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/captcha"
	"regexp"
	"strings"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
)

type AccountController struct {
	BaseController
}

var cpt *captcha.Captcha

func init() {
	// 通过使用beego缓存系统来存储二维码数据
	// cache.FileCache 是 Beego 框架提供的一种缓存类型，用于将数据存储在文件系统中。
	// 二维码数据将被存储在相对于当前工作目录的 ./cache/captcha 目录中。
	fc := &cache.FileCache{CachePath: "/captcha"}

	//captcha.Captcha.NewWithFilter方法创建一个验证码对象 cpt。该方法接受两个参数，
	//第一个参数是二维码路由的前缀，这里设置为 "/captcha/"。
	//第二个参数是缓存对象，这里传入之前创建的 fc。
	//通过这个方法，将创建一个带有路由前缀和缓存系统的验证码对象。
	cpt = captcha.NewWithFilter("/captcha", fc)
}

// 登录
func (c *AccountController) Login() {
	var remember CookieRemember

	// 验证cookie
	// GetSecureCookie 从已编码的浏览器 cookie 值返回已解码的 cookie 值。
	// 采用了sha256来作为加密算法，
	// 第一个参数secret是加密的密钥，第二个参数key是cookie的名字，返回cookie值
	if cookie, ok := c.GetSecureCookie(common.AppKey(), "login"); ok {
		// 如果cookie可以解码到remember结构体中
		if err := utils.Decode(cookie, &remember); err == nil {
			if err = c.login(remember.MemberId); err == nil {
				// 重定向页面
				c.Redirect(web.URLFor("HomeController.Index"), 302)
				return
			}
		}
	}
	// 如果当前请求是get请求，那么直接展示页面即可
	c.TplName = "account/login.html"

	// 当前请求是否为Post请求，如果是就执行登录过程
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

		// 设置两个session
		c.SetMember(*member)

		remember.MemberId = member.MemberId
		remember.Account = member.Account
		remember.Time = time.Now()
		v, err := utils.Encode(remember)
		if err == nil {
			c.SetSecureCookie(common.AppKey(), "login", v, 20*3600*365)
		}
		c.JsonResult(0, "ok")
	}

	c.Data["RandomStr"] = time.Now().Unix()
}

// 注册页面
func (c *AccountController) Regist() {
	var (
		nickname  string      // 昵称
		avatar    string      //头像的http链接地址
		email     string      // 邮箱地址
		username  string      // 用户名
		id        interface{} // 用户id
		captchaOn bool        // 是否开启了验证码
	)

	// 如果开启了验证码
	//strings.EqualFold(v, "true") 和 v == "true"的区别：
	//strings.EqualFold(v, "true")不区分大小写
	//例如，"true" 和 "True"、"TRuE" 等都会被认为是相等的。
	if v, ok := c.Option["ENABLED_CAPTCHA"]; ok && strings.EqualFold(v, "true") {
		// 如果开启了验证码，将captchaOn设置为true
		captchaOn = true
		c.Data["CaptchaOn"] = captchaOn
	}

	c.Data["NickName"] = nickname
	c.Data["Avatar"] = avatar
	c.Data["Email"] = email
	c.Data["Username"] = username
	c.Data["Id"] = id
	c.Data["RandomStr"] = time.Now().Unix()
	// 存储标识，已标记是哪个用户。在完善用户信息的时候跟传递过来的auth和id进行校验
	// c.SetSession 方法用于设置会话的键值对。在这里，代码将键 "auth" 的值设置为 "email-id"
	c.SetSession("auth", fmt.Sprintf("%v-%v", "email", id))
	c.TplName = "account/bind.html"
}

// DoRegist 处理注册时提交的注册信息
func (c *AccountController) DoRegist() {
	var err error
	account := c.GetString("account")
	//strings.TrimSpace用于去除字符串首尾的空白字符（包括空格、制表符、换行符等）
	nickname := strings.TrimSpace(c.GetString("nickname"))
	password1 := c.GetString("password1")
	password2 := c.GetString("password2")
	email := c.GetString("email")

	member := models.NewMember()

	if password1 != password2 {
		c.JsonResult(1, "登录密码与确认密码不一致")
	}

	//使用 strings.Count 函数计算字符串 password1 中的字符个数（不包括空字符）
	// strings.Count(password1, "")将始终返回字符串 password1 的长度加 1
	if l := strings.Count(password1, "") - 1; password1 == "" || l > 20 || l < 6 {
		c.JsonResult(1, "密码必须在6-20个字符之间")
	}

	// 正则表达式匹配，邮箱需要满足一定的格式
	if ok, err := regexp.MatchString(common.RegexpEmail, email); !ok || err != nil {
		c.JsonResult(1, "邮箱格式错误")
	}

	if l := strings.Count(nickname, "") - 1; l < 2 || l > 20 {
		c.JsonResult(1, "用户昵称限制在2-20个字符")
	}

	member.Account = account
	member.Nickname = nickname
	member.Password = password1

	if account == "admin" || account == "administrator" {
		// 超级管理员
		member.Role = common.MemberSuperRole
	} else {
		// 普通用户
		member.Role = common.MemberGeneralRole
	}
	// 默认头像
	member.Avatar = common.DefaultAvatar()
	member.CreateAt = 0
	member.Email = email
	member.Status = 0

	if err := member.Add(); err != nil {
		logs.Error(err.Error())
		c.JsonResult(1, err.Error())
	}

	// 如果注册的数据没有问题，将自动登录
	if err = c.login(member.MemberId); err != nil {
		logs.Error(err.Error())
		c.JsonResult(1, err.Error())
	}

	c.JsonResult(0, "注册成功")
}

// 退出登录
func (c *AccountController) Logout() {
	// 设置当前会员对象为空
	c.SetMember(models.Member{})
	// 设置安全 Cookie 的值为空字符串，并将过期时间设置为 -3600，即立即过期
	c.SetSecureCookie(common.AppKey(), "login", "", -3600)
	// 重定向到登录页面，302 表示临时重定向。这样，用户在注销后会被重定向到登录页面。
	c.Redirect(web.URLFor("AccountController.Login"), 302)
}

/*
* 私有函数
 */
//封装一个内部调用的函数，login
// 判断数据库是否有相应的memberId
func (c *AccountController) login(memberId int) (err error) {
	//能否通过该memberId查询到相应用户
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
	if err == nil {
		c.SetSecureCookie(common.AppKey(), "login", v, 20*3600*365)
	}
	return err
}
