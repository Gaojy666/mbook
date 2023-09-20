package models

import "github.com/beego/beego/v2/client/orm"

func init() {
	// 注册所有的表，以便orm框架可以访问
	orm.RegisterModel(
		new(Category),
		new(Book),
		new(BookCategory),
	)
}

// 定义几个表名，通过函数来返回
// 将表名写进函数的原因：是平时开发的原则，开发过程中如果需要改表名，可以在一个地方统一地改
func TNCategory() string {
	return "md_category"
}

func TNBook() string {
	return "md_book"
}
func TNBookCategory() string {
	return "md_book_category"
}
