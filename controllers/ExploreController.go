package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"math"
	"strconv"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
)

type ExploreController struct {
	BaseController
}

func (c *ExploreController) Index() {
	var (
		cid  int // 分类id
		cate models.Category
		// 超过一页，动态生成一些链接，链接的前缀是相同的，一般是域名
		// 这里取的是index对应的路径，只是为了分页而服务
		urlPrefix = web.URLFor("ExploreController.Index")
	)

	//// GetInt用来获取URL中?后面携带的参数
	//if cid, _ := c.GetInt("cid"); cid > 0 {
	//	// 取到分类号后，将对应的category的信息拿出来
	//	cateModel := new(models.Category)
	//	cate = cateModel.Find(cid)
	//	c.Data["Cate"] = cate
	//}

	cidstr := c.Ctx.Input.Param(":cid")
	if len(cidstr) > 0 {
		if cid, _ = strconv.Atoi(cidstr); cid > 0 {
			cateModel := new(models.Category)
			cate = cateModel.Find(cid)
			c.Data["Cate"] = cate
		}
	}

	c.Data["Cid"] = cid
	c.TplName = "explore/index.html"

	// 获取图书信息

	// 分页操作，当图书多的时候可能有分页
	// 如果没有分页，则默认值设为1，表示从第1页去取
	pageIndex, _ := c.GetInt("page", 1)
	pageSize := 24

	// 取出分类下所有的图书数据
	books, totalCount, err := models.NewBook().HomeData(pageIndex, pageSize, cid)
	if err != nil {
		logs.Error(err)
		c.Abort("404")
	}

	// 将cid放到url中
	if totalCount > 0 {
		urlSuffix := ""
		if cid > 0 {
			urlSuffix = urlSuffix + "&cid=" + strconv.Itoa(cid)
		}
		html := utils.NewPaginations(4, totalCount, pageSize, pageIndex, urlPrefix, urlSuffix)
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}

	// 计算总页数，ceil向上取整
	c.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	c.Data["Lists"] = books
}
