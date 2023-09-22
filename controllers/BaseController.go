package controllers

import (
	"compress/gzip"
	"encoding/json"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"io"
	"strings"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
)

type BaseController struct {
	web.Controller
}

type CookieRemember struct {
	MemberId int
	Account  string
	Time     time.Time
}

// 设置登录用户信息
func (c *BaseController) SetMember(member models.Member) {
	if member.MemberId <= 0 {
		// 删除会话数据
		c.DelSession(common.SessionName)
		c.DelSession("uid")
		c.DestroySession()
	} else {
		c.SetSession(common.SessionName, member)
		c.SetSession("uid", member.MemberId)
	}
}

// Ajax接口返回Json
// JsonResult将错误码、错误消息和数据组装成 JSON 格式的响应并返回给客户端
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
