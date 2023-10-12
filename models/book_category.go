package models

import (
	"strconv"
)

type BookCategory struct {
	Id         int // 自增主键
	BookId     int // book id
	CategoryId int // 分类 id
}

func (m *BookCategory) TableName() string {
	return TNBookCategory()
}

// 根据书籍id查询分类id
func (m *BookCategory) SelectByBookId(book_id int) (cates []Category, rows int64, err error) {
	o := GetOrm("r")
	sql := "select c.* from " + TNCategory() + " c left join " + TNBookCategory() + " bc on c.id=bc.category_id where bc.book_id=?"
	rows, err = o.Raw(sql, book_id).QueryRows(&cates)
	return
}

// 处理书籍分类
func (m *BookCategory) SetBookCates(bookId int, cids []string) {
	if len(cids) == 0 {
		return
	}

	var (
		cates             []Category
		tableCategory     = TNCategory()
		tableBookCategory = TNBookCategory()
	)

	o := GetOrm("w")
	o.QueryTable(tableCategory).Filter("id__in", cids).All(&cates, "id", "pid")

	//创建了一个空的 cidMap，用于存储分类ID和其父分类ID的映射关系。
	cidMap := make(map[string]bool)
	for _, cate := range cates {
		//遍历 cates，将分类ID和其父分类ID存入 cidMap 中
		cidMap[strconv.Itoa(cate.Pid)] = true
		cidMap[strconv.Itoa(cate.Id)] = true
	}
	//将cid及其父类id全部放入新的 cids 数组中
	cids = []string{}
	for cid, _ := range cidMap {
		cids = append(cids, cid)
	}

	// 将原book_id对应book的种类删除
	o.QueryTable(tableBookCategory).Filter("book_id", bookId).Delete()
	var bookCates []BookCategory
	for _, cid := range cids {
		cidNum, _ := strconv.Atoi(cid)
		bookCate := BookCategory{
			CategoryId: cidNum,
			BookId:     bookId,
		}
		bookCates = append(bookCates, bookCate)
	}
	if l := len(bookCates); l > 0 {
		// 向book_category表中添加新的分类数据
		o.InsertMulti(l, &bookCates)
	}
	//开启一个goroutine,进行分类计数的更新操作。
	go CountCategory()
}
