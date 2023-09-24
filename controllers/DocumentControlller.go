package controllers

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"strings"
	"ziyoubiancheng/mbook/models"
)

type DocumentController struct {
	BaseController
}

// 图书目录&详情页
func (c *DocumentController) Index() {
	// 从json请求中获取私有图书的访问token
	token := c.GetString("token")
	// 从路由参数中获取名为":key"的值
	identify := c.Ctx.Input.Param(":key")
	if identify == "" {
		// 如果":key"为空，终止请求并返回404错误
		c.Abort("404")
	}
	// 从json中获取键为tab的值，并借其转换为小写
	tab := strings.ToLower(c.GetString("tab"))

	// 1. 获取图书详情
	bookResult := c.getBookData(identify, token)
	if bookResult.BookId == 0 {
		// 没有阅读权限,重定向到主页
		c.Redirect(web.URLFor("HomeController.Index"), 302)
		return
	}

	c.TplName = "document/intro.html"
	c.Data["Book"] = bookResult

	switch tab {
	case "comment", "score":
		// 如果tab的值为 "comment" 或 "score"，则为空，后期可添加内容

	default:
		tab = "default"
	}
	c.Data["Tab"] = tab
	// 2. 获取该书的目录
	c.Data["Menu"], _ = new(models.Document).GetMenuTop(bookResult.BookId)

	// 3. 获取其他用户对该书的评论内容，包括评分，用户昵称，评论内容，头像，评论时间
	c.Data["Comments"], _ = new(models.Comments).BookComments(1, 30, bookResult.BookId)
	// 用户自己对该书的评分
	c.Data["MyScore"] = new(models.Score).BookScoreByUid(c.Member.MemberId, bookResult.BookId)
}

// 阅读器页面
func (c *DocumentController) Read() {
	// Read需要取两个内容，一个是目录内容，一个是详情内容

	identify := c.Ctx.Input.Param(":key")
	id := c.GetString(":id")
	token := c.GetString("token")

	if identify == "" || id == "" {
		c.Abort("404")
	}

	// 没开启匿名
	if !c.EnableAnonymous && c.Member == nil {
		c.Redirect(web.URLFor("AccountController.Login"), 302)
		return
	}

	// 如果有权限，拿到本书的相关信息
	bookData := c.getBookData(identify, token)

	doc := models.NewDocument()
	// 将某章节的内容拿出来
	doc, err := doc.SelectByIdentify(bookData.BookId, id) // 文档标识
	if err != nil {
		c.Abort("404")
	}

	if doc.BookId != bookData.BookId {
		c.Abort("404")
	}

	// 对章节内容做一些渲染
	if doc.Release != "" {
		// 使用了 goquery 包来解析 HTML 文档并创建一个 Document 对象 query
		// doc.Release 是一个字符串，表示 HTML 文档的内容。
		query, err := goquery.NewDocumentFromReader(bytes.NewBufferString(doc.Release))
		if err != nil {
			logs.Error(err)
		} else {
			// query 对象调用 Find 方法，根据选择器 "img" 查找匹配的 DOM 元素。
			// 这里查找了所有的 <img> 元素。
			// 而后.Each()方法对每个 <img> 元素执行操作。
			query.Find("img").Each(func(i int, contentSelection *goquery.Selection) {
				// 是否能获取 <img> 元素的 src 属性值。
				if _, ok := contentSelection.Attr("src"); ok {

				}
				// 获取 <img> 元素的 alt 属性值，判断alt属性是否为空
				if alt, _ := contentSelection.Attr("alt"); alt == "" {
					// 设置 <img> 元素的 alt 属性为指定的值。
					// 这里使用 doc.DocumentName 和索引值 i+1 构造了一个新的 alt 属性值。
					contentSelection.SetAttr("alt", doc.DocumentName+" - 图"+fmt.Sprint(i+1))
				}
			})
			//通过 query 对象调用 Find 方法，根据选择器 "body" 查找匹配的 DOM 元素，
			//并使用 Html 方法获取其 HTML 内容。
			html, err := query.Find("body").Html()
			if err != nil {
				logs.Error(err)
			} else {
				doc.Release = html
			}
		}
	}

	// 将章节的附件也取出来
	attach, err := models.NewAttachment().SelectByDocumentId(doc.DocumentId)
	if err != nil {
		doc.AttachList = attach
	}

	// 更新数据库数据
	// 1.图书阅读人次+1
	if err := models.IncOrDec(models.TNBook(), "vcnt",
		fmt.Sprintf("book_id=%v", doc.BookId),
		true, 1,
	); err != nil {
		logs.Error(err.Error())
	}

	// 2.文档阅读人次+1
	if err := models.IncOrDec(models.TNDocuments(), "vcnt",
		fmt.Sprintf("document_id=%v", doc.DocumentId),
		true, 1,
	); err != nil {
		logs.Error(err.Error())
	}

	doc.Vcnt = doc.Vcnt + 1

	// 处理Ajax请求
	if c.IsAjax() {
		var data struct {
			Id        int    `json:"doc_id"`
			DocTitle  string `json:"doc_title"`  // 章节标题
			Body      string `json:"body"`       // 主体内容
			Title     string `json:"title"`      //
			View      int    `json:"view"`       // 视图
			UpdatedAt string `json:"updated_at"` // 更新日期
		}
		data.DocTitle = doc.DocumentName
		data.Body = doc.Release
		data.Id = doc.DocumentId
		data.View = doc.Vcnt
		data.UpdatedAt = doc.ModifyTime.Format("2006-01-02 15:04:05")

		c.JsonResult(0, "ok", data)
	}
}

// 获取图书内容并判断权限
func (c *DocumentController) getBookData(identify, token string) *models.BookData {
	// 根据图书的唯一标识，去查询图书内容
	book, err := models.NewBook().Select("identify", identify)
	if err != nil {
		logs.Error(err)
		c.Abort("404")
	}

	// 私有文档 并且 用户不是管理员
	if book.PrivatelyOwned == 1 && !c.Member.IsAdministrator() {
		isOk := false
		if c.Member != nil {
			// 可以根据bookId和memberId查询到该用户是一名管理员
			_, err := models.NewRelationship().SelectRoleId(book.BookId, c.Member.MemberId)
			if err == nil {
				isOk = true
			}
		}
		// 如果该书有私有token，并且用户不是管理员
		if book.PrivateToken != "" && !isOk {
			// 比较传入的token和数据库中私有图书的token
			if token != "" && strings.EqualFold(token, book.PrivateToken) {
				// 如果相同，设置name为图书identify的session
				c.SetSession(identify, token)
			} else if token, ok := c.GetSession(identify).(string); !ok || !strings.EqualFold(token, book.PrivateToken) {
				// 如果令牌不相同，或者没有传入令牌，
				// 则检查会话中存储的令牌与数据库中的令牌是否相同，如果不相同则返回 404。
				c.Abort("404")
			}
		} else if !isOk {
			//如果图书没有私有访问令牌，并且用户也不是管理员，则返回 404。
			c.Abort("404")
		}
	}

	// 将book类型转换为BookData类型
	bookResult := book.ToBookData()
	if c.Member != nil {
		// 根据图书ID和用户ID查询关联关系，并将结果存储到 bookResult 中的相应字段。
		rsh, err := models.NewRelationship().Select(bookResult.BookId, c.Member.MemberId)
		if err == nil {
			bookResult.MemberId = rsh.MemberId
			bookResult.RoleId = rsh.RoleId
			bookResult.RelationshipId = rsh.RelationshipId
		}
	}
	return bookResult
}
