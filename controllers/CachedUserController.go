package controllers

import (
	"github.com/beego/beego/v2/server/web"
	"log"
	"strconv"
	"ziyoubiancheng/mbook/common"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
	"ziyoubiancheng/mbook/utils/dynamicache"
)

type CachedUserController struct {
	BaseController
	UcenterMember models.Member
}

func (c *CachedUserController) Prepare() {
	// 初始化用户登录信息
	c.BaseController.Prepare()

	username := c.GetString(":username")
	// 从缓存中读取用户信息
	cachekeyUser := "dynamicache_user:" + username
	err := dynamicache.ReadStruct(cachekeyUser, &c.UcenterMember)
	if err != nil {
		// 缓存中没有，查数据库
		c.UcenterMember, _ = new(models.Member).GetByUsername(username)
		// 写缓存
		dynamicache.WriteStruct(cachekeyUser, c.UcenterMember)
	}

	if c.UcenterMember.MemberId == 0 {
		c.Abort("404")
		return
	}
	c.Data["IsSelf"] = c.UcenterMember.MemberId == c.Member.MemberId
	log.Printf("****-----------%d\n", c.UcenterMember.MemberId == c.Member.MemberId)
	c.Data["User"] = c.UcenterMember
	c.Data["Tab"] = "share"
}

func (c *CachedUserController) Index() {
	page, _ := c.GetInt("page")
	pageSize := 10
	if page < 1 {
		page = 1
	}

	//动态缓存读取c.Data["Books"]信息
	var books []*models.BookData
	cachekeyBookList := "dynamicache_userbook_" + strconv.Itoa(c.UcenterMember.MemberId) + "_page_" + strconv.Itoa(page)
	// 先尝试从缓存中读取
	totalCount, err := dynamicache.ReadList(cachekeyBookList, &books)
	if err != nil {
		// 缓存中没有，查数据库，并把结果写入缓存
		books, totalCount, _ = models.NewBook().SelectPage(page, pageSize, c.UcenterMember.MemberId, 0)
		dynamicache.WriteList(cachekeyBookList, books, totalCount)
	}

	c.Data["Books"] = books

	if totalCount > 0 {
		html := utils.NewPaginations(common.RollPage, totalCount, pageSize, page, web.URLFor("CachedUserController.Index", ":username", c.UcenterMember.Account), "")
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["Total"] = totalCount
	c.TplName = "user/index.html"
}

func (c *CachedUserController) Collection() {
	page, _ := c.GetInt("page")
	pageSize := 10
	if page < 1 {
		page = 1
	}

	// 读取c.Data["Books"]信息
	var books []models.CollectionData
	var totalCount int64
	cachekeyCollectionList := "dynamicache_usercollection " + strconv.Itoa(c.UcenterMember.MemberId) + "_page_" + strconv.Itoa(page)
	total, err := dynamicache.ReadList(cachekeyCollectionList, &books)
	totalCount = int64(total)
	if err != nil {
		totalCount, books, _ = new(models.Collection).List(c.UcenterMember.MemberId, page, pageSize)
		dynamicache.WriteList(cachekeyCollectionList, books, int(totalCount))
	}
	c.Data["Books"] = books

	if totalCount > 0 {
		html := utils.NewPaginations(common.RollPage, int(totalCount), pageSize, page, web.URLFor("CachedUserController.Collection", ":username", c.UcenterMember.Account), "")
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["Total"] = totalCount
	c.Data["Tab"] = "collection"
	c.TplName = "user/collection.html"

}

func (c *CachedUserController) Follow() {
	page, _ := c.GetInt("page")
	pageSize := 18
	if page < 1 {
		page = 1
	}

	// 读取关注列表缓存
	var fans []models.FansData
	var totalCount int64
	cachekeyFollowList := "dynamicache_userfollow " + strconv.Itoa(c.UcenterMember.MemberId) + "_page_" + strconv.Itoa(page)
	total, err := dynamicache.ReadList(cachekeyFollowList, &fans)
	totalCount = int64(total)
	if err != nil {
		fans, totalCount, _ = new(models.Fans).FollowList(c.UcenterMember.MemberId, page, pageSize)
		dynamicache.WriteList(cachekeyFollowList, fans, int(totalCount))
	}

	if totalCount > 0 {
		html := utils.NewPaginations(common.RollPage, int(totalCount), pageSize, page, web.URLFor("CachedUserController.Follow", ":username", c.UcenterMember.Account), "")
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["Fans"] = fans
	c.Data["Tab"] = "follow"
	c.TplName = "user/fans.html"
}

func (c *CachedUserController) Fans() {
	page, _ := c.GetInt("page")
	pageSize := 18
	if page < 1 {
		page = 1
	}

	var fans []models.FansData
	var totalCount int64
	cachekeyFanList := "dynamicache_userfans_ " + strconv.Itoa(c.UcenterMember.MemberId) + "_page_" + strconv.Itoa(page)
	total, err := dynamicache.ReadList(cachekeyFanList, &fans)
	totalCount = int64(total)
	if err != nil {
		fans, totalCount, _ = new(models.Fans).FansList(c.UcenterMember.MemberId, page, pageSize)
		dynamicache.WriteList(cachekeyFanList, fans, int(totalCount))
	}
	if totalCount > 0 {
		html := utils.NewPaginations(common.RollPage, int(totalCount), pageSize, page, web.URLFor("CachedUserController.Fans", ":username", c.UcenterMember.Account), "")
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["Fans"] = fans
	c.Data["Tab"] = "fans"
	c.TplName = "user/fans.html"
}
