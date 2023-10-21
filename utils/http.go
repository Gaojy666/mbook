package utils

import (
	"errors"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/bitly/go-simplejson"
	"io"
	"time"
)

// Put请求，传递给ES新建索引
func HttpPutJson(url, body string) error {
	// 为请求体添加请求头，发送请求，接收响应
	response, err := httplib.Put(url).
		Header("Content-Type", "application/json").
		SetTimeout(10*time.Second, 10*time.Second).Body(body).Response()
	if err == nil {
		defer response.Body.Close()
		// http收发是正常的，但是内容是错的
		if response.StatusCode >= 300 || response.StatusCode < 200 {
			body, _ := io.ReadAll(response.Body)
			err = errors.New(response.Status + ";" + string(body))
		}
	}
	return err
}

// Post请求，搜索
// 将response的内容解析成json返回
func HttpPostJson(url, body string) (*simplejson.Json, error) {
	response, err := httplib.Post(url).
		Header("Content-Type", "application/json").
		SetTimeout(10*time.Second, 10*time.Second).Body(body).Response()
	var sj *simplejson.Json
	if err == nil {
		defer response.Body.Close()
		if response.StatusCode >= 300 || response.StatusCode < 200 {
			body, _ := io.ReadAll(response.Body)
			err = errors.New(response.Status + "; " + string(body))
		} else {
			// 接收的response是正常的
			bodyBytes, _ := io.ReadAll(response.Body)
			// simplejson就是go的json的封装
			// 包装了一个结构体，通过其库的方法可以往指定的地方加一些节点或读一些节点
			// 用起来比较方便
			sj, err = simplejson.NewJson(bodyBytes)
		}
	}

	return sj, err
}
