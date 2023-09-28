package controllers

type SearchController struct {
	BaseController
}

// 搜索首页
func (c *SearchController) Search() {
	c.TplName = "search/search.html"
}

// 搜索结果页
func (c *SearchController) Result() {

}
