package dynamicache

import (
	"encoding/json"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"time"
)

var (
	pool      *redis.Pool = nil
	MaxIdle   int         = 0
	MaxOpen   int         = 0
	ExpireSec int64       = 0
)

// 初始化redis连接池
func InitCache() {
	addr, _ := web.AppConfig.String("dynamicache_addrstr")
	if len(addr) == 0 {
		addr = "127.0.0.1:6379"
	}
	if MaxIdle <= 0 {
		MaxIdle = 256
	}
	pass, _ := web.AppConfig.String("dynamicache_passwd")
	if len(pass) == 0 {
		pool = &redis.Pool{
			MaxIdle:     MaxIdle,
			MaxActive:   MaxOpen,
			IdleTimeout: time.Duration(120),
			Dial: func() (redis.Conn, error) {
				return redis.Dial(
					"tcp",
					addr,
					redis.DialReadTimeout(1*time.Second),
					redis.DialWriteTimeout(1*time.Second),
					redis.DialConnectTimeout(1*time.Second),
				)
			},
		}
	} else {
		pool = &redis.Pool{
			MaxIdle:     MaxIdle,
			MaxActive:   MaxOpen,
			IdleTimeout: time.Duration(120),
			Dial: func() (redis.Conn, error) {
				return redis.Dial(
					"tcp",
					addr,
					redis.DialReadTimeout(1*time.Second),
					redis.DialWriteTimeout(1*time.Second),
					redis.DialConnectTimeout(1*time.Second),
					redis.DialPassword(pass),
				)
			},
		}
	}
}

func rdsdo(cmd string, key interface{}, args ...interface{}) (interface{}, error) {
	con := pool.Get()
	if err := con.Err(); err != nil {
		return nil, err
	}

	// 将key和args拼接成整个一个切片
	params := make([]interface{}, 0)
	params = append(params, key)

	if len(args) > 0 {
		for _, v := range args {
			params = append(params, v)
		}
	}

	return con.Do(cmd, params...)
}

func WriteString(key string, value string) error {
	_, err := rdsdo("SET", key, value)
	logs.Debug("redis set:" + key + "-" + value)
	// 设置过期时间
	rdsdo("EXPIRE", key, ExpireSec)
	return err
}

func ReadString(key string) (string, error) {
	result, err := rdsdo("GET", key)
	logs.Debug("redis get:" + key)
	if err == nil {
		// 因为result是空接口类型，因此需要转成字符串类型
		// 如果传进去的err不为空，则返回""
		str, _ := redis.String(result, err)
		return str, nil
	} else {
		logs.Debug("redis get error:" + err.Error())
		return "", err
	}
}

// 结构化数据封装，将obj转化成字符串，而后调用writeString方法
// 相当于在redis中存储了结构化的字符串进去
func WriteStruct(key string, obj interface{}) error {
	// 先将obj转化成字符串
	data, err := json.Marshal(obj)
	if err == nil {
		return WriteString(key, string(data))
	} else {
		// json没有转成功
		return err
	}
}

// 按照传入的obj结构，将value放入到obj中,传入的是obj的引用
func ReadStruct(key string, obj interface{}) error {
	if data, err := ReadString(key); err == nil {
		return json.Unmarshal([]byte(data), obj)
	} else {
		return err
	}
}

// 由于社区界面有分页，因此还需要传入总数据数，以便分页
func WriteList(key string, list interface{}, total int) error {
	// 将list数据和tatal分别写入两个key中
	realKeyList := key + "_list"
	realKeyCount := key + "_count"
	data, err := json.Marshal(list)
	if err == nil {
		WriteString(realKeyCount, strconv.Itoa(total))
		WriteString(realKeyList, string(data))
		return nil
	}
	return err
}

func ReadList(key string, list interface{}) (int, error) {
	//realKeyList := key + "_list"
	//realKeyCount := key + "_count"
	//
	//data, err := ReadString(realKeyList)
	//if err != nil {
	//	return 0, err
	//}
	//json.Unmarshal([]byte(data), list)
	//count, err := ReadString(realKeyCount)
	//if err != nil {
	//	return 0, err
	//}
	//total, _ := strconv.Atoi(count)
	//return total, nil

	realKeyList := key + "_list"
	realKeyCount := key + "_count"
	if data, err := ReadString(realKeyList); nil == err {
		totalStr, _ := ReadString(realKeyCount)
		total := 0
		if len(totalStr) > 0 {
			total, _ = strconv.Atoi(totalStr)
		}
		return total, json.Unmarshal([]byte(data), list)
	} else {
		return 0, err
	}
}
