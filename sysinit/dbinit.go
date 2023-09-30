package sysinit

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	// 先调用models层中的init函数，创建表
	_ "ziyoubiancheng/mbook/models"
)

// dbinit() // 初始化主库
// dbinit("w") 或 dbinit("default")
// dbinit("w", "r"...)
func dbinit(aliases ...string) {
	// 如果在开发模式下，需要输出debug信息，isDev判断是否是开发模式
	Dev, _ := web.AppConfig.String("runmode")
	isDev := (Dev == "dev")

	if len(aliases) > 0 {
		for _, alias := range aliases {
			RegisterDatabase(alias)
			if alias == "w" {
				// 若是主库，则自动建表
				// 注册了一个名为 "default" 的 MySQL 数据库连接
				// false 表示只在表不存在时创建，isDev 表示在开发模式下，输出详细的同步操作日志。
				orm.RunSyncdb("default", false, isDev)
			}
		}
	} else {
		// 无参初始化主库的情况
		RegisterDatabase("w")
		// 自动建表
		orm.RunSyncdb("default", false, isDev)
	}

	if isDev {
		orm.Debug = isDev
	}
}

// 注册单库
func RegisterDatabase(alias string) {
	if len(alias) <= 0 {
		return
	}

	// 根据alias拼接字符串，确定是主数据库还是从数据库
	// 默认是主数据库

	dbAlias := alias // default,最少要有一个default,其他的可以指定

	// 如果为主库
	// 为了程序的健壮性，防止别人将alias认为是dbAlias的作用
	// 因此要做一个兼容
	if alias == "w" || alias == "default" || len(alias) <= 0 {
		// dbAlias是建立数据库连接时传进连接函数的参数
		dbAlias = "default"
		// 拼接连接字符串的参数
		alias = "w"
	}

	// 拼接字符串
	// 数据库名称
	dbName, _ := web.AppConfig.String("db_" + alias + "_database")
	// 数据库用户名
	dbuser, _ := web.AppConfig.String("db_" + alias + "_username")
	// 数据库密码
	dbPwd, _ := web.AppConfig.String("db_" + alias + "_password")
	// 数据库IP
	dbHost, _ := web.AppConfig.String("db_" + alias + "_host")
	// 数据库端口
	dbPost, _ := web.AppConfig.String("db_" + alias + "_port")

	//ORM 必须注册一个别名为 default 的数据库，作为默认使用
	// root:lu741208@tcp(127.0.0.1:3306)/mbook?charset=utf8mb4
	orm.RegisterDataBase(dbAlias, "mysql",
		dbuser+":"+dbPwd+"@tcp("+dbHost+":"+dbPost+")/"+
			dbName+"?charset=utf8mb4", orm.MaxIdleConnections(30))
}
