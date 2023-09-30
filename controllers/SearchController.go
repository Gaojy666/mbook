package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"time"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
)

type SearchController struct {
	BaseController
}

// 搜索首页
func (c *SearchController) Search() {
	c.TplName = "search/search.html"
}

// 搜索结果页
func (c *SearchController) Result() {
	totalRows := 0
	var ids []int

	wd := c.GetString("wd")
	if "" == wd {
		c.Redirect(web.URLFor("SearchController.Search"), 302)
	}

	now := time.Now()

	tab := c.GetString("tab", "doc") // 默认按章节内容查询
	isSearchDoc := false
	if tab == "doc" {
		isSearchDoc = true
	}
	page, _ := c.GetInt("page", 1) // 默认第一页

	if page < 1 {
		page = 1
	}
	size := 10

	if isSearchDoc { // 按文档搜索
		docs, count, err := models.NewDocumentSearch().SearchDocument(wd, 0, page, size)
		totalRows = count
		if err != nil {
			logs.Error(err.Error())
		} else {
			for _, doc := range docs {
				ids = append(ids, doc.DocumentId)
			}
		}
	} else {
		// 按书籍搜索
		books, count, err := models.NewBook().SearchBook(wd, page, size)
		totalRows = count
		if err != nil {
			logs.Error(err.Error())
		} else {
			for _, book := range books {
				// 拿到所有查询到的书籍id
				ids = append(ids, book.BookId)
			}
		}
	}

	if len(ids) > 0 {
		if isSearchDoc {
			c.Data["Docs"], _ = models.NewDocumentSearch().GetDocsById(ids)
		} else {
			c.Data["Books"], _ = models.NewBook().GetBooksByIds(ids)
		}
	}

	c.Data["totalRows"] = totalRows
	if totalRows > size {
		// 有分页
		if totalRows > 1000 {
			totalRows = 1000
		}
		urlSuffix := fmt.Sprintf("&tab=%v&wd=%v", tab, wd)
		// 得到分页的html
		html := utils.NewPaginations(4, totalRows, size, page, web.URLFor("SearchController.Result"), urlSuffix)
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}

	c.Data["SpendTime"] = fmt.Sprintf("%.3f", time.Since(now).Seconds())
	c.Data["Wd"] = wd
	c.Data["Tab"] = tab
	c.TplName = "search/result.html"
}
