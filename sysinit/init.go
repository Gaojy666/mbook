package sysinit

// init()在main函数之前只会被调用一次
// init()函数不能有参数，不能有返回值
func init() {
	sysinit()
	dbinit("w") // 初始化主库
	dbinit("r") // 初始化从库
	dbinit("uar")
	dbinit("uaw")
}
