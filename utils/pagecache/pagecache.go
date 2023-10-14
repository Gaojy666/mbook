package pagecache

import (
	"context"
	"errors"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"strings"
	"time"
)

var (
	BasePath  string           = ""  // 缓存路径
	ExpireSec int64            = 0   // 缓存过期时间
	store     *cache.FileCache = nil // 初始化一个filecache对象
	// 将pagecache.conf中的pagecache_list放大map里面，key是controller和action
	cacheMap map[string]bool     = nil
	paramMap map[string][]string = nil
)

func InitCache() {
	store = &cache.FileCache{CachePath: BasePath}
	pagecacheList, _ := web.AppConfig.Strings("pagecache_list")

	// 初始化静态化配置列表
	cacheMap = make(map[string]bool)
	for _, v := range pagecacheList {
		cacheMap[strings.ToLower(v)] = true
	}

	paramMap = make(map[string][]string)
	// 分类页面缓存
	pagecacheMap, _ := web.AppConfig.GetSection("pagecache_param")
	for k, v := range pagecacheMap {
		// key: ExploreController_Index, v: :cid:bid:kid ...
		sv := strings.Split(v, ";")
		paramMap[k] = sv
	}
}

// 判断controller和action是否在缓存list中，从而判断是否要都缓存
func IncacheList(controllerName, actionName string) bool {
	keyname := cacheKey(controllerName, actionName)
	if f := cacheMap[keyname]; f {
		return f
	}
	return false
}

// 判断在Prepare函数中是否要写缓存,超时的话再去写
func NeedWrite(controllerName, actionName string, params map[string]string) bool {
	// IncacheList判断在写缓存列表中
	if IncacheList(controllerName, actionName) {
		keyName := cacheKey(controllerName, actionName, params)
		// 判断是否超时通过get函数
		getCache, _ := store.Get(context.Background(), keyName)

		tmpCache, ok := getCache.(string)
		// 如果tmpCache字符串为空，说明超时
		if !ok && len(tmpCache) == 0 {
			logs.Debug("need write :" + keyName)
			return true
		}
	}
	return false
}

// 写缓存
// content传指针，因为content比较大，防止形成大的值拷贝
func Write(controllerName, actionName string, content *string, params map[string]string) error {
	keyname := cacheKey(controllerName, actionName, params)
	if len(keyname) == 0 {
		return errors.New("未找到缓存key")
	}

	err := store.Put(context.Background(), keyname, *content, time.Duration(ExpireSec)*time.Second)
	return err
}

// 读缓存
func Read(controllerName, actionName string, params map[string]string) (*string, error) {
	keyname := cacheKey(controllerName, actionName, params)
	if len(keyname) == 0 {
		return nil, errors.New("未找到缓存key")
	}

	tmpContent, _ := store.Get(context.Background(), keyname)

	//tmp, _ := store.IsExist(context.Background(), keyname)
	//fmt.Println(tmp)

	content, ok := tmpContent.(string)
	if ok {
		return &content, nil
	}
	return nil, errors.New("缓存无法找到或格式错误")
}

// 通过Controller name和action name来生成一个缓存用到的key
func cacheKey(controllerName, actionName string, paramArray ...map[string]string) string {
	if len(controllerName) > 0 && len(actionName) > 0 {
		rtnstr := strings.ToLower(controllerName + "_" + actionName)
		if len(paramArray) > 0 {
			for _, v := range paramMap[rtnstr] {
				// v = :cid
				// ExploreController_index_cid_1
				rtnstr = rtnstr + "_" + strings.ReplaceAll(v, ":", "") + "_" + paramArray[0][v]
				//fmt.Println(paramArray[0][v])
			}
		}
		return rtnstr
	}
	return ""
}

// 清理过期缓存文件，通过定时任务调用
func ClearExpiredFiles() {
	for k, _ := range cacheMap {
		ok, _ := store.IsExist(context.Background(), k)
		if ok {
			tmpContent, _ := store.Get(context.Background(), k)
			content, ok := tmpContent.(string)
			if !ok && len(content) == 0 {
				// 缓存过期，将过期缓存删除
				store.Delete(context.Background(), k)
			}
		}
	}
}
