package controllers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/session"
	_ "github.com/beego/beego/v2/server/web/session/redis"
	"io"
	"log"
	"strings"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
	"ziyoubiancheng/mbook/utils/RabbitMQ"
	"ziyoubiancheng/mbook/utils/pagecache"
)

type BaseController struct {
	web.Controller
	Member          *models.Member    // 用户
	Option          map[string]string // 全局设置
	EnableAnonymous bool              // 开启匿名访问
	Rabbitmq        *RabbitMQ.RabbitMQ
}

type CookieRemember struct {
	MemberId int
	Account  string
	Time     time.Time
}

var globalSessions *session.Manager
var Rabbitmq *RabbitMQ.RabbitMQ

func init() {

	web.BConfig.WebConfig.Session.SessionOn, _ = web.AppConfig.Bool("sessionon")
	web.BConfig.WebConfig.Session.SessionProvider = "redis"
	web.BConfig.WebConfig.Session.SessionProviderConfig = "127.0.0.1:6379"

	sessionCfg, _ := web.AppConfig.String("sessionproviderconfig")
	if len(sessionCfg) == 0 {
		return
	}
	// 1.创建一个session配置
	sessionConfig := &session.ManagerConfig{
		CookieName:      "gosessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  sessionCfg,
	}
	// 2.创建全局的session管理对象
	var err error
	globalSessions, err = session.NewManager("redis", sessionConfig)
	if err != nil {
		log.Printf(err.Error())
	}
	// 3.开启协程，负责session垃圾回收,定时1小时清理过期的session数据
	go globalSessions.GC()
}

func (c *BaseController) Finish() {
	controllerName, actionName := c.GetControllerAndAction()
	if pagecache.NeedWrite(controllerName, actionName, c.Ctx.Input.Params()) {
		// 拿出模板渲染后的字符串
		render, err := c.RenderString()
		//fmt.Println(render)
		if len(render) > 0 && err == nil {
			// write render to file 写到对应的缓存文件中去
			err = pagecache.Write(controllerName, actionName, &render, c.Ctx.Input.Params())
		}
	}
}

// 每个子类Controller公用方法调用前，都执行一下Prepare方法
func (c *BaseController) Prepare() {
	// 定义全局rabbitmq对象
	Rabbitmq = RabbitMQ.NewRabbitMQSimple("imoocProduct")

	//如果有缓存，则返回缓存内容
	controllerName, actionName := c.GetControllerAndAction()
	if pagecache.IncacheList(controllerName, actionName) {
		contentPtr, err := pagecache.Read(controllerName, actionName, c.Ctx.Input.Params())
		// 返回的不是空字符串，说明没有过期
		if err == nil && len(*contentPtr) > 0 {
			// 给用户返回缓存的内容, 下面的查询数据库操作就不执行了
			io.WriteString(c.Ctx.ResponseWriter, *contentPtr)
			logs.Debug(controllerName + "-" + actionName + "read Cache")
			c.StopRun()
		}
	}

	c.Member = models.NewMember() // 初始化
	c.EnableAnonymous = false

	// 从redis的session中获取用户信息
	// 将session放入redis是为了配置nginx后多服务器的session同步问题

	if globalSessions == nil {
		c.Abort("500")
	}
	//根据当前请求返回 session 对象
	session, err := globalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
	if err != nil {
		logs.Debug(err.Error())
	}
	defer session.SessionRelease(context.Background(), c.Ctx.ResponseWriter) //释放session中的资源，保存数据

	// 判断当前member是否在session中
	memberInSession := false
	var member models.Member
	memberobj := session.Get(context.Background(), common.SessionName)
	if memberobj != nil {
		member = memberobj.(models.Member)
		if member.MemberId > 0 {
			memberInSession = true
		}
	}

	//if member, ok := c.GetSession(common.SessionName).(models.Member); ok && member.MemberId > 0 {
	if memberInSession {
		c.Member = &member
	} else {
		// 如果Session中没有检测到
		// 如果Cookie中存在登录信息，从Cookie中获取memberId,然后在数据库中查找对应的用户信息
		if cookie, ok := c.GetSecureCookie(common.AppKey(), "login"); ok {
			var remember CookieRemember
			err := utils.Decode(cookie, &remember)
			if err == nil {
				member, err := models.NewMember().Find(remember.MemberId)
				if err == nil {
					c.SetMember(*member)
					c.Member = member
				}
			}
		}
	}

	if c.Member.RoleName == "" {
		c.Member.RoleName = common.Role(c.Member.Role)
	}
	// 返回前端需要的信息
	c.Data["Member"] = c.Member
	c.Data["BaseUrl"] = c.BaseUrl()
	c.Data["SITE_NAME"] = "MBOOK"
	// 设置全局配置
	c.Option = make(map[string]string)
	c.Option["ENABLED_CAPTCHA"] = "false"
}

// 设置登录用户信息
func (c *BaseController) SetMember(member models.Member) {
	//if member.MemberId <= 0 {
	//	// 删除会话数据
	//	c.DelSession(common.SessionName)
	//	c.DelSession("uid")
	//	c.DestroySession()
	//} else {
	//	// 如果用户信息存在，将其写入到session中
	//	// 默认session以文件的形式
	//	c.SetSession(common.SessionName, member)
	//	c.SetSession("uid", member.MemberId)
	//}
	session, _ := globalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
	defer session.SessionRelease(context.Background(), c.Ctx.ResponseWriter)
	if member.MemberId <= 0 {
		session.Delete(context.Background(), common.SessionName)
		session.Delete(context.Background(), "uid")
	} else {
		session.Set(context.Background(), common.SessionName, member)
		session.Set(context.Background(), "uid", member.MemberId)
	}
}

// Ajax接口返回Json
// 将错误码、错误消息和数据组装成 JSON 格式的响应并返回给客户端
func (c *BaseController) JsonResult(errCode int, errMsg string, data ...interface{}) {
	jsonData := make(map[string]interface{}, 3)
	jsonData["errcode"] = errCode
	jsonData["message"] = errMsg

	if len(data) > 0 && data[0] != nil {
		jsonData["data"] = data[0]
	}
	returnJSON, err := json.Marshal(jsonData)
	if err != nil {
		logs.Error(err)
	}
	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 如果请求头包含了"gzip"编码，启用gzip压缩
	if strings.Contains(strings.ToLower(c.Ctx.Request.Header.Get("Accept-Encoding")), "gzip") {
		c.Ctx.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		//创建一个 Gzip 编码的写入器（gzip.Writer）
		//将压缩后的 JSON 数据写入响应，并在结束后关闭写入器。
		w := gzip.NewWriter(c.Ctx.ResponseWriter)
		defer w.Close()
		w.Write(returnJSON)
		w.Flush()
	} else {
		io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))
	}
	//调用 StopRun 方法来停止当前请求的执行
	c.StopRun()
}

// 应该是返回配置项中设置的另一台主机的地址
func (c *BaseController) BaseUrl() string {
	// sitemap_host 什么意思？
	host, _ := web.AppConfig.String("sitemap_host")
	if len(host) > 0 {
		// 检查 host 是否以 "http://" 或 "https://" 开头
		if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
			return host
		}
		//如果 host 不包含完整的 URL 地址，
		//使用 c.Ctx.Input.Scheme() 获取请求的协议（HTTP 或 HTTPS），
		//然后与 host 组合成完整的 URL 地址，并返回。
		return c.Ctx.Input.Scheme() + "://" + host
	}
	//如果没有获取到有效的 "sitemap_host" 配置项的值
	//使用 c.Ctx.Input.Scheme() 获取请求的协议（HTTP 或 HTTPS），
	//再加上 c.Ctx.Request.Host 获取当前请求的主机名，组合成完整的 URL 地址
	return c.Ctx.Input.Scheme() + "://" + c.Ctx.Request.Host
}

// 关注或取消关注
func (c *BaseController) SetFollow() {
	if c.Member.MemberId == 0 {
		c.JsonResult(1, "请先登录")
	}
	uid, _ := c.GetInt(":uid")
	if uid == c.Member.MemberId {
		c.JsonResult(1, "不能关注自己")
	}
	cancel, _ := new(models.Fans).FollowOrCancel(uid, c.Member.MemberId)
	if cancel {
		c.JsonResult(0, "已成功取消关注")
	}
	c.JsonResult(0, "已成功关注")
}
