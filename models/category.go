package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"strings"
)

type Category struct {
	Id     int
	Pid    int    // 父ID,pid=0时为一级分类，二级分类和一级分类的关系为多对一，因此形成自连表
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

	//首先按照 "status" 字段的降序排序（使用负号表示降序），显示排在前面隐藏排在后面
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

// 更新分类字段
func (m *Category) UpdateField(id int, field, val string) (err error) {
	_, err = orm.NewOrm().QueryTable(TNCategory()).Filter("id", id).Update(orm.Params{field: val})
	return
}

// 统计分类书籍
var counting = false

type Count struct {
	Cnt        int
	CategoryId int
}

// 分类计数的更新操作
func CountCategory() {
	//表示有其他goroutine在进行计数操作，直接返回，避免重复计数
	if counting {
		return
	}
	counting = true
	defer func() {
		counting = false
	}()

	var count []Count

	o := orm.NewOrm()
	//查询分类的计数结果
	sql := "select count(bc.id) cnt, bc.category_id from " + TNBookCategory() + " bc left join " + TNBook() + " b on b.book_id=bc.book_id where b.privately_owned=0 group by bc.category_id"
	o.Raw(sql).QueryRows(&count)
	// 没有计数结果,直接返回
	if len(count) == 0 {
		return
	}

	var cates []Category
	//查询所有分类的信息
	o.QueryTable(TNCategory()).All(&cates, "id", "pid", "cnt")
	// 没有查询到,直接返回
	if len(cates) == 0 {
		return
	}

	var err error

	to, _ := o.Begin()
	defer func() {
		if err != nil {
			//如果在更新过程中发生错误，立即返回，回滚事务。
			to.Rollback()
		} else {
			//如果计数过程中没有发生错误，则提交事务。
			to.Commit()
		}
	}()

	//所有分类的计数置为0
	to.QueryTable(TNCategory()).Update(orm.Params{"cnt": 0})
	//空的 cateChild 字典，用于存储分类ID和对应的子分类数量。
	cateChild := make(map[int]int)
	for _, item := range count {
		if item.Cnt > 0 {
			// 更新分类的计数
			cateChild[item.CategoryId] = item.Cnt
			_, err = to.QueryTable(TNCategory()).Filter("id", item.CategoryId).Update(orm.Params{"cnt": item.Cnt})
			if err != nil {
				return
			}
		}
	}
}

// 删除分类
func (m *Category) Delete(id int) (err error) {
	var cate = Category{Id: id}

	o := orm.NewOrm()
	if err = o.Read(&cate); cate.Cnt > 0 { //当前分类下文档图书数量不为0，不允许删除
		return errors.New("删除失败，当前分类下的问下图书不为0，不允许删除")
	}

	if _, err = o.Delete(&cate, "id"); err != nil {
		return
	}
	// 将处于该分类下的子分类也删除掉
	_, err = o.QueryTable(TNCategory()).Filter("pid", id).Delete()
	if err != nil { //删除分类图标
		return
	}

	return
}

// 批量新增分类
func (m *Category) InsertMulti(pid int, cates string) (err error) {
	//按照换行符 \n 进行分割
	slice := strings.Split(cates, "\n")
	if len(slice) == 0 {
		return
	}

	o := orm.NewOrm()
	for _, item := range slice {
		if item = strings.TrimSpace(item); item != "" {
			var cate = Category{
				Pid:    pid,
				Title:  item,
				Status: true,
			}
			if o.Read(&cate, "title"); cate.Id == 0 {
				// 数据库中上没有该名称的分类
				_, err = o.Insert(&cate)
			}
		}
	}
	return
}
