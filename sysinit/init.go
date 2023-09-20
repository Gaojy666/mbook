package sysinit

// init()在main函数之前只会被调用一次
// init()函数不能有参数，不能有返回值
func init() {
	sysinit()
	dbinit()
}
