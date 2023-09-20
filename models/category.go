package models

import "github.com/beego/beego/v2/client/orm"

type Category struct {
	Id     int
	Pid    int    // 分类ID
	Title  string `orm:"size(30);unique"`
	Intro  string // 介绍
	Icon   string // 回调路径？
	Cnt    int    // 统计分类下图书
	Sort   int    // 排序
	Status bool   // 状态，true显示,false隐藏
}

func (m *Category) TableName() string {
	return TNCategory()
}

// 获取所有分类
func (m *Category) GetCates(pid int, status int) (cates []Category, err error) {
	qs := orm.NewOrm().QueryTable(TNCategory())
	// 如果想得到全部的分类，那么pid传进来是-1
	if pid > -1 {
		qs = qs.Filter("pid", pid)
	}
	if status == 0 || status == 1 {
		qs = qs.Filter("status", status)
	}

	//首先按照 "status" 字段的降序排序（使用负号表示降序），
	//然后按照 "sort" 字段的升序排序，最后按照 "title" 字段的升序排序。
	//All() 方法将查询结果作为切片类型传递给 &cates
	_, err = qs.OrderBy("-status", "sort", "title").All(&cates)
	return
}

// 通过Category_id,来获得分类信息
func (m *Category) Find(cid int) (cate Category) {
	cate.Id = cid
	orm.NewOrm().Read(&cate)
	return cate
}
