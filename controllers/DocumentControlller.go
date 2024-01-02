package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils/dynamicache"
	"ziyoubiancheng/mbook/utils/store"
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
	//c.Data["Menu"], _ = new(models.Document).GetMenuTop(bookResult.BookId)

	// 3. 获取其他用户对该书的评论内容，包括评分，用户昵称，评论内容，头像，评论时间
	// 当前默认展示30条评论
	//c.Data["Comments"], _ = new(models.Comments).BookComments(1, 30, bookResult.BookId)
	// 用户自己对该书的评分
	c.Data["MyScore"] = new(models.Score).BookScoreByUid(c.Member.MemberId, bookResult.BookId)

	// 动态缓存逻辑

	// 创建独一的key
	cachekeyDocidx := "dynamicache_document_index_cdata:" + identify

	// 动态缓存c.Data["Menu"]
	cachekeyDocidxMenu := cachekeyDocidx + "_menu"
	var dataMenu []models.Document
	// （1）先尝试取读缓存，如果读到直接返回
	err := dynamicache.ReadStruct(cachekeyDocidxMenu, &dataMenu)
	if err != nil {
		// （2）没有读到，再查询数据库
		dataMenu, _ = new(models.Document).GetMenuTop(bookResult.BookId)
		// (3) 并写缓存
		dynamicache.WriteStruct(cachekeyDocidxMenu, dataMenu)
	}
	c.Data["Menu"] = dataMenu

	cachekeyDocidxComments := cachekeyDocidx + "_comments"
	var dataComments []models.BookCommentsResult
	err = dynamicache.ReadStruct(cachekeyDocidxComments, &dataComments)
	if err != nil {
		dataComments, _ = new(models.Comments).BookComments(1, 30, bookResult.BookId)
		dynamicache.WriteStruct(cachekeyDocidxComments, dataComments)
		logs.Error(err.Error())
	}

	c.Data["Comments"] = dataComments
}

// 阅读器页面
func (c *DocumentController) Read() {
	// Read需要取两个内容，一个是章节目录内容，一个是章节详情内容

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

	// 1.拿到章节详情内容
	// 如果有权限，拿到本书的相关信息
	bookData := c.getBookData(identify, token)

	doc := models.NewDocument()

	// 将某章节的内容拿出来(在从库中进行查询)
	doc, err := doc.SelectByIdentify(bookData.BookId, id) // 文档标识
	//docId, _ := strconv.Atoi(id)
	//doc, err := doc.SelectByDocId(docId)
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

		// 1. 返回章节详情内容
		c.JsonResult(0, "ok", data)
	}

	// 2. 拿到本书的章节目录，同时选中的当前章节，目录中要有高亮
	// tree取得返回的由菜单树结构生成的html字符串
	tree, err := models.NewDocument().GetMenuHtml(bookData.BookId, doc.DocumentId)

	if err != nil {
		logs.Error(err)
		c.Abort("404")
	}

	c.Data["Bookmark"] = false
	c.Data["Model"] = bookData
	c.Data["Book"] = bookData
	c.Data["Result"] = template.HTML(tree) // 返回菜单
	c.Data["Title"] = doc.DocumentName
	c.Data["DocId"] = doc.DocumentId
	c.Data["Content"] = template.HTML(doc.Release)
	c.Data["View"] = doc.Vcnt
	c.Data["UpdatedAt"] = doc.ModifyTime.Format("2006-01-02 15:04:05")

	//设置模版
	c.TplName = "document/default_read.html"
}

// 读书编辑页面
// 先展示一个模板，尚未填充内容
func (c *DocumentController) Edit() {
	docId := 0 // 文档id

	identify := c.Ctx.Input.Param(":key")
	if identify == "" {
		c.Abort("404")
	}

	bookData := models.NewBookData()

	var err error
	//现根据用户进行权限验证
	if c.Member.IsAdministrator() {
		// 根据"identify"找到对应的book数据
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			c.JsonResult(1, "权限错误")
		}
		// 将book类型转为BookData类型
		bookData = book.ToBookData()
	} else {
		// 如果不是管理员，再根据书和用户的关系再进行权限认证
		bookData, err = models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)
		if err != nil {
			c.Abort("404")
		}
		// 普通用户不能打开该书的编辑页
		if bookData.RoleId == common.BookGeneral {
			c.JsonResult(1, "权限错误")
		}
	}

	// 渲染模板
	c.TplName = "document/markdown_edit_template.html"

	c.Data["Model"] = bookData
	r, _ := json.Marshal(bookData)

	c.Data["ModelResult"] = template.JS(string(r))

	c.Data["Result"] = template.JS("[]")

	// 编辑的文档
	if id := c.GetString(":id"); id != "" {
		if num, _ := strconv.Atoi(id); num > 0 {
			docId = num
		} else { //字符串 or num <= 0
			var doc = models.NewDocument()
			// 查到符合指定identify和book_id的那条document的id
			models.GetOrm("w").QueryTable(doc).Filter("identify", id).Filter("book_id", bookData.BookId).One(doc, "document_id")
			docId = doc.DocumentId
		}
	}

	// 取到章节菜单目录，并且当前选中的章节要高亮
	// 并且将当前选中的章节设置为编辑模式
	trees, err := models.NewDocument().GetMenu(bookData.BookId, docId, true)
	if err != nil {
		logs.Error("GetMenu error : ", err)
	} else {
		if len(trees) > 0 {
			if jsTree, err := json.Marshal(trees); err == nil {
				c.Data["Result"] = template.JS(string(jsTree))
			}
		} else {
			c.Data["Result"] = template.JS("[]")
		}
	}
	c.Data["BaiDuMapKey"] = web.AppConfig.DefaultString("baidumapkey", "")
}

// 向Edit编辑页面填充内容
// 如果是Post请求，就保存文档并返回内容
func (c *DocumentController) Content() {
	// 获取请求参数
	identify := c.Ctx.Input.Param(":key")
	docId, err := c.GetInt("doc_id")
	errMsg := "ok"
	// 根据请求参数获取章节Id
	if err != nil {
		docId, _ = strconv.Atoi(c.Ctx.Input.Param(":id"))
	}
	bookId := 0

	// 权限验证
	if c.Member.IsAdministrator() {
		// 如果用户是管理员，则根据标识符获取书籍信息
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			c.JsonResult(1, "获取内容错误")
		}
		bookId = book.BookId
	} else {
		// 如果用户不是管理员，则根据标识符和用户ID获取书籍数据信息
		bookData, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)

		// 普通用户没有编辑该书权限
		if err != nil || bookData.RoleId == common.BookGeneral {
			c.JsonResult(1, "权限错误")
		}
		bookId = bookData.BookId
	}

	// 检查参数是否正确
	if docId <= 0 {
		c.JsonResult(1, "参数错误")
	}

	// documentStore是保存功能中重点操作的对象
	documentStore := new(models.DocumentStore)

	// 如果不是POST请求，则返回文档内容
	if !c.Ctx.Input.IsPost() {
		// 根据docId取到章节所有信息
		doc, err := models.NewDocument().SelectByDocId(docId)

		if err != nil {
			c.JsonResult(1, "文档不存在")
		}
		// 取章节中的附件
		attach, err := models.NewAttachment().SelectByDocumentId(doc.DocumentId)
		if err == nil {
			doc.AttachList = attach
		}

		doc.Release = "" //Ajax请求，之间用markdown渲染，不用release
		doc.Markdown = documentStore.SelectField(doc.DocumentId, "markdown")
		c.JsonResult(0, errMsg, doc)
	}

	//更新文档内容
	//请求中获取 markdown 字段的值，去除首尾空白字符
	markdown := strings.TrimSpace(c.GetString("markdown", ""))
	//请求中获取 html 字段的值
	content := c.GetString("html")

	//从请求中获取 version 和 cover 字段的值，并尝试将 version 转换为 int64 类型。
	version, _ := c.GetInt64("version", 0)
	isCover := c.GetString("cover")

	doc, err := models.NewDocument().SelectByDocId(docId)

	if err != nil {
		c.JsonResult(1, "读取文档错误")
	}
	if doc.BookId != bookId {
		c.JsonResult(1, "内部错误")
	}
	if doc.Version != version && !strings.EqualFold(isCover, "yes") {
		c.JsonResult(1, "文档将被覆盖")
	}

	isSummary := false
	isAuto := false

	// 只有markdown没有指定并且content指定时,返回html格式给documentStore.Markdown
	if markdown == "" && content != "" {
		documentStore.Markdown = content
	} else {
		documentStore.Markdown = markdown
	}
	documentStore.Content = content
	doc.Version = time.Now().Unix()

	// 插入或更新文档
	if docId, err := doc.InsertOrUpdate(); err != nil {
		c.JsonResult(1, "保存失败")
	} else {
		documentStore.DocumentId = int(docId)
		if err := documentStore.InsertOrUpdate("markdown", "content"); err != nil {
			logs.Error(err)
		}
	}

	if isAuto {
		errMsg = "auto"
	} else if isSummary {
		errMsg = "true"
	}

	doc.Release = ""
	c.JsonResult(0, errMsg, doc)
}

// 阅读页内搜索
func (c *DocumentController) Search() {
	identify := c.Ctx.Input.Param(":key")
	token := c.GetString("token")
	keyword := strings.TrimSpace(c.GetString("keyword"))

	if identify == "" {
		c.JsonResult(1, "参数错误")
	}
	if !c.EnableAnonymous && c.Member == nil {
		c.Redirect(web.URLFor("AccountController.Login"), 302)
		return
	}
	bookData := c.getBookData(identify, token)
	docs, _, err := models.NewDocumentSearch().SearchDocument(keyword, bookData.BookId, 1, 10000)
	if err != nil {
		logs.Error(err)
		c.JsonResult(1, "搜索结果错误")
	}
	c.JsonResult(0, keyword, docs)
}

// 上传附件
func (c *DocumentController) Upload() {
	identify := c.GetString("identify")
	docId, _ := c.GetInt("doc_id")
	isAttach := true

	if identify == "" {
		c.JsonResult(1, "参数错误")
	}
	name := "editormd-file-file"
	//尝试从请求中获取名为 "editormd-file-file" 的文件对象
	//如果该文件字段存在并成功获取到文件对象，
	//那么 file 变量将表示文件对象，
	//moreFile 变量将表示更多的文件信息（如文件名、大小等）
	file, moreFile, err := c.GetFile(name)
	if err == http.ErrMissingFile {
		// 如果上传的不是文件,是图片
		name = "editormd-image-file"
		file, moreFile, err = c.GetFile(name)
		// 如果都没有获取到,则返回错误
		if err == http.ErrMissingFile {
			c.JsonResult(1, "文件错误")
		}
	}
	if err != nil {
		c.JsonResult(1, err.Error())
	}

	defer file.Close()

	//获取文件的扩展名，并进行相关的错误检查。
	ext := filepath.Ext(moreFile.Filename)
	if ext == "" {
		c.JsonResult(1, "文件格式错误")
	}

	if !common.IsAllowedFileExt(ext) {
		c.JsonResult(1, "文件类型错误")
	}

	bookId := 0
	//如果是管理员，则不判断权限
	if c.Member.IsAdministrator() {
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			c.JsonResult(1, "文档不存在或权限不足")
		}
		bookId = book.BookId
	} else {
		book, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)
		if err != nil {
			if err == orm.ErrNoRows {
				c.JsonResult(1, "权限错误")
			}
			c.JsonResult(6001, err.Error())
		}
		//没有编辑权限
		if book.RoleId != common.BookEditor && book.RoleId != common.BookAdmin && book.RoleId != common.BookFounder {
			c.JsonResult(1, "权限错误")
		}
		bookId = book.BookId
	}

	if docId > 0 {
		doc, err := models.NewDocument().SelectByDocId(docId)
		if err != nil {
			c.JsonResult(1, "获取文档错误")
		}
		if doc.BookId != bookId {
			c.JsonResult(1, "获取文档错误")
		}
	}

	//根据当前时间生成文件名，并构建文件的保存路径。
	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)
	//common.WorkingDirectory：一个变量或常量，表示工作目录的路径。
	//"uploads"：一个字符串，表示文件上传目录的名称。
	//time.Now().Format("200601")：使用当前时间生成一个格式为 "200601" 的字符串，表示年份和月份。这将作为文件上传目录的子目录。
	//fileName+ext：文件名加上文件扩展名，构成最终的文件名。
	filePath := filepath.Join(common.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName+ext)
	//提取 filePath 的目录部分，将其赋值给变量 path
	path := filepath.Dir(filePath)

	//创建文件保存目录，并设置权限
	//os.MkdirAll() 函数会递归地创建目录，如果目录已存在，则不会报错。它会根据提供的路径创建缺失的目录，确保最终路径中的所有目录都存在。
	//os.ModePerm 是一个常量，表示目录的权限。它指定了权限位，使得创建的目录具有读、写和执行权限（即 0777）
	os.MkdirAll(path, os.ModePerm)

	//将文件保存到指定路径
	err = c.SaveToFile(name, filePath)

	if err != nil {
		c.JsonResult(1, "保存文件失败")
	}
	attachment := models.NewAttachment()
	attachment.BookId = bookId
	attachment.Name = moreFile.Filename
	attachment.CreateAt = c.Member.MemberId
	attachment.Ext = ext
	attachment.Path = strings.TrimPrefix(filePath, common.WorkingDirectory)
	attachment.DocumentId = docId

	//os.Stat() 函数来获取文件的信息
	//fileInfo 是一个变量，用于接收 os.Stat() 函数的返回值，
	//其中包含了文件的各种属性，例如文件大小、修改时间等
	if fileInfo, err := os.Stat(filePath); err == nil {
		attachment.Size = float64(fileInfo.Size())
	}
	if docId > 0 {
		attachment.DocumentId = docId
	}

	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") || strings.EqualFold(ext, ".png") || strings.EqualFold(ext, ".gif") {

		//将文件路径转换为相对于工作目录的相对路径，并将路径中的反斜杠替换为正斜杠。
		//strings.TrimPrefix(filePath, common.WorkingDirectory) 去除文件路径中工作目录的前缀部分，得到相对于工作目录的路径。
		//strings.Replace() 函数将路径中的反斜杠 \ 替换为正斜杠 /，以确保路径在不同操作系统下的兼容性
		//在路径前添加 /，使其成为以根目录为起点的绝对路径
		attachment.HttpPath = "/" + strings.Replace(strings.TrimPrefix(filePath, common.WorkingDirectory), "\\", "/", -1)
		if strings.HasPrefix(attachment.HttpPath, "//") {
			attachment.HttpPath = string(attachment.HttpPath[1:])
		}
		isAttach = false
	}

	err = attachment.Insert()

	if err != nil {
		os.Remove(filePath)
		c.JsonResult(1, "文件保存失败")
	}
	if attachment.HttpPath == "" {
		//设置附件的 HttpPath 为下载附件的链接，并将附件信息更新到数据库中。
		//"DocumentController.DownloadAttachment" 表示控制器名为 DocumentController 的 DownloadAttachment 方法。
		//":key", identify, ":attach_id", attachment.AttachmentId 是一组键值对形式的路由参数。
		//这些参数将根据路由规则进行替换，生成最终的 URL 路径。
		attachment.HttpPath = web.URLFor("DocumentController.DownloadAttachment", ":key", identify, ":attach_id", attachment.AttachmentId)

		if err := attachment.Update(); err != nil {
			c.JsonResult(1, "保存文件失败")
		}
	}
	//根据项目标识和文件名构建 OSS 存储路径，并保存文件到 OSS 存储。
	//filepath.Ext(attachment.HttpPath) 函数用于获取 attachment.HttpPath 中的文件扩展名部分。
	osspath := fmt.Sprintf("projects/%v/%v", identify, fileName+filepath.Ext(attachment.HttpPath))

	osspath = "uploads/" + osspath
	//OSS（对象存储服务）通常是指阿里云提供的云端对象存储服务
	if err := store.SaveToLocal("."+attachment.HttpPath, osspath); err != nil {
		logs.Error(err.Error())
	}
	//将附件的 HttpPath 设置为 OSS 存储的访问路径。
	attachment.HttpPath = "/" + osspath

	result := map[string]interface{}{
		"errcode":   0,
		"success":   1,
		"message":   "ok",
		"url":       attachment.HttpPath,
		"alt":       attachment.Name,
		"is_attach": isAttach,
		"attach":    attachment,
	}
	c.Ctx.Output.JSON(result, true, false)
	c.StopRun()
}

// 创建文档
func (c *DocumentController) Create() {
	identify := c.GetString("identify")        //图书标识
	docIdentify := c.GetString("doc_identify") //新建的文档标识
	docName := c.GetString("doc_name")
	parentId, _ := c.GetInt("parent_id", 0)
	docId, _ := c.GetInt("doc_id", 0)
	bookIdentify := strings.TrimSpace(c.GetString(":key"))
	o := models.GetOrm("w") // 使用别名"w"来选择数据库
	if identify == "" {
		c.JsonResult(1, "参数错误")
	}
	if docName == "" {
		c.JsonResult(1, "文档名为空")
	}
	if docIdentify != "" {
		if bookIdentify == "" {
			c.JsonResult(1, "图书参数错误")
		}

		var book models.Book
		o.QueryTable(models.TNBook()).Filter("Identify", bookIdentify).One(&book, "BookId")
		if book.BookId == 0 {
			c.JsonResult(1, "未找到该图书")
		}

		d, _ := models.NewDocument().SelectByIdentify(book.BookId, docIdentify)
		if d.DocumentId > 0 && d.DocumentId != docId {
			// 两个不同的章节id对应同一个章节,是不成立的
			c.JsonResult(1, "文档标识重复")
		}
	} else {
		// 使用当前日期时间生成一个默认的docIdentify
		docIdentify = fmt.Sprintf("date-%v", time.Now().Format("2019.11.02.01.01.05"))
	}

	bookId := 0
	if c.Member.IsAdministrator() {
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			logs.Error(err)
			// 这里不应该是权限错误,存疑
			c.JsonResult(1, "权限错误")
		}
		bookId = book.BookId
	} else {
		bookData, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)

		// 普通用户没有添加章节的权限
		if err != nil || bookData.RoleId == common.BookGeneral {
			c.JsonResult(1, "权限错误")
		}
		bookId = bookData.BookId
	}

	if parentId > 0 {
		doc, err := models.NewDocument().SelectByDocId(parentId)
		// 父章节和当前章节一定在同一本书里
		if err != nil || doc.BookId != bookId {
			c.JsonResult(1, "分类错误")
		}
	}

	document, _ := models.NewDocument().SelectByDocId(docId)

	document.MemberId = c.Member.MemberId
	document.BookId = bookId
	if docIdentify != "" {
		document.Identify = docIdentify
	}
	document.Version = time.Now().Unix()
	document.DocumentName = docName
	document.ParentId = parentId

	documentId, err := document.InsertOrUpdate()
	if err != nil {
		c.JsonResult(1, "保存失败")
	}

	documentStore := models.DocumentStore{DocumentId: int(documentId), Markdown: ""}
	if documentStore.SelectField(documentId, "markdown") == "" {
		if err := documentStore.InsertOrUpdate(); err != nil {
			logs.Error(err)
		}
	}
	c.JsonResult(0, "ok", document)
}

// 删除
func (c *DocumentController) Delete() {

	identify := c.GetString("identify")
	docId, _ := c.GetInt("doc_id", 0)

	bookId := 0
	if c.Member.IsAdministrator() {
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			c.JsonResult(1, "权限错误")
		}
		bookId = book.BookId
	} else {
		bookData, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)
		if err != nil || bookData.RoleId == common.BookGeneral {
			c.JsonResult(1, "权限错误")
		}
		bookId = bookData.BookId
	}

	if docId <= 0 {
		c.JsonResult(1, "参数错误")
	}

	doc, err := models.NewDocument().SelectByDocId(docId)
	if err != nil {
		c.JsonResult(1, "删除失败")
	}

	//如果文档所属图书错误
	if doc.BookId != bookId {
		c.JsonResult(1, "参数错误")
	}
	//删除图书下的文档以及子文档
	err = doc.Delete(doc.DocumentId)
	if err != nil {
		logs.Error(err.Error())
		c.JsonResult(1, "删除失败")
	}

	//文档数量统计
	models.NewBook().RefreshDocumentCount(doc.BookId)

	c.JsonResult(0, "ok")
}

// 获取图书内容和用户间的关系, 若是私有图书判断权限
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
