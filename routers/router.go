package routers

import (
	"github.com/beego/beego/v2/server/web"
	"ziyoubiancheng/mbook/controllers"
)

func init() {
	//web.Router("/", &controllers.MainController{})
	// 首页&分类
	// 通过get方法请求到controller中的Index方法里面
	web.Router("/", &controllers.HomeController{}, "get:Index")
	web.Router("/2", &controllers.HomeController{}, "get:Index2")
	web.Router("/explore", &controllers.ExploreController{}, "get:Index")
	web.Router("/books/:key", &controllers.DocumentController{}, "get:Index")

	// 读书模块
	web.Router("/read/:key/:id", &controllers.DocumentController{}, "*:Read")         // 图书目录&详情
	web.Router("/read/:key/search", &controllers.DocumentController{}, "post:Search") // 图书章节内容搜索

	// 图书编辑模块
	web.Router("/api/:key/edit/?:id", &controllers.DocumentController{}, "*:Edit")       // 文档编辑
	web.Router("/api/:key/content/?:id", &controllers.DocumentController{}, "*:Content") // 保存文档内容
	web.Router("/api/upload", &controllers.DocumentController{}, "post:Upload")          // 上传图片
	web.Router("/api/:key/create", &controllers.DocumentController{}, "post:Create")     // 创建章节
	web.Router("/api/:key/delete", &controllers.DocumentController{}, "post:Delete")     // 删除章节

	// 搜索  ******
	//web.Router("/search", &controllers.SearchController{}, "get:Search")
	//web.Router("/search/result", &controllers.SearchController{}, "get:Result")

	// login
	// 与regist相比，login没有类似doregist的请求，其post请求在AccountController通过判断
	// 如果是post请求，就执行登录过程；如果是get请求，就展示get页面
	web.Router("/login", &controllers.AccountController{}, "*:Login")
	// Register页面
	web.Router("/regist", &controllers.AccountController{}, "*:Regist")
	web.Router("/logout", &controllers.AccountController{}, "*:Logout")
	// 注册时提交的post请求
	web.Router("/doregist", &controllers.AccountController{}, "post:DoRegist")

	//用户图书管理
	web.Router("/book", &controllers.BookController{}, "*:Index")                         //我的图书
	web.Router("/book/create", &controllers.BookController{}, "post:Create")              //创建图书
	web.Router("/book/:key/setting", &controllers.BookController{}, "*:Setting")          //图书设置
	web.Router("/book/setting/upload", &controllers.BookController{}, "post:UploadCover") //图书封面
	web.Router("/book/star/:id", &controllers.BookController{}, "*:Collection")           //收藏图书
	web.Router("/book/setting/save", &controllers.BookController{}, "post:SaveBook")      //保存
	web.Router("/book/:key/release", &controllers.BookController{}, "post:Release")       //发布
	web.Router("/book/setting/token", &controllers.BookController{}, "post:CreateToken")  //创建Token

	// 个人中心  ******
	web.Router("/user/:username", &controllers.UserController{}, "get:Index")            // 分享
	web.Router("/user/:username/collection", &controllers.UserController{}, "get:Index") // 收藏
	web.Router("/user/:username/follow", &controllers.UserController{}, "get:Follow")    // 关注
	web.Router("/user/:username/fans", &controllers.UserController{}, "get:Fans")        // 粉丝
	web.Router("/follow/:uid", &controllers.BaseController{}, "get:SetFollow")           // 关注或取消关注
	web.Router("/book/score/:id", &controllers.BookController{}, "*:Score")              // 评分
	web.Router("/book/comment/:id", &controllers.BookController{}, "post:Comment")       // 评论

	// 个人设置
	// /setting中，如果是get请求，展示一个setting页面；如果是post请求，会把建好的信息给入库
	web.Router("/setting", &controllers.SettingController{}, "*:Index")         // 用户界面
	web.Router("/setting/upload", &controllers.SettingController{}, "*:Upload") // 上传头像

	// 管理后台
	web.Router("/manager/category", &controllers.ManagerController{}, "post,get:Category")    // 分类管理首页
	web.Router("/manager/update-cate", &controllers.ManagerController{}, "get:UpdateCate")    // 更新分类信息
	web.Router("/manager/del-cate", &controllers.ManagerController{}, "get:DelCate")          // 删除分类信息
	web.Router("/manager/icon-cate", &controllers.ManagerController{}, "post:UpdateCateIcon") // 更新图标
}
