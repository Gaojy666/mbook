package utils

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	html1 "html/template"
	"os"
	"path/filepath"
	"strings"
)

//存储类型

// 更多存储类型有待扩展
const (
	Version           = "1.0"
	StoreLocal string = "local"
	StoreOss   string = "oss"
)

var (
	BasePath, _  = filepath.Abs(filepath.Dir(os.Args[0]))
	StoreType, _ = web.AppConfig.String("store_type") //存储类型
)

// 评分处理
func ScoreFloat(score int) string {
	return fmt.Sprintf("%1.1f", float32(score)/10.0)
}

// 操作图片显示
// 如果用的是oss存储，这style是avatar、cover可选项
func ShowImg(img string, style ...string) (url string) {
	//路径以 "https://" 或 "http://" 开头，则直接返回该图片路径，表示图片已经是完整的 URL。
	if strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "http://") {
		return img
	}
	//路径不是以 "https://" 或 "http://" 开头，则认为图片存储在本地文件系统中，需要进行处理
	img = "/" + strings.TrimLeft(img, "./")
	switch StoreType {
	//图片存储在 OSS（对象存储服务）中。
	case StoreOss:
		s := ""
		//如果函数调用时提供了 style 参数，表示需要对图片进行样式处理，例如 "avatar" 或 "cover"
		if len(style) > 0 && strings.TrimSpace(style[0]) != "" {
			s = "/" + style[0]
		}
		//从应用配置中获取 OSS 的域名配置，并将其赋值给 url 变量
		url, _ = web.AppConfig.String("oss::Domain")
		//去除 url 变量末尾的斜杠 "/" 和空格。
		//将图片路径、样式参数和 url 拼接起来，形成完整的图片 URL
		url = strings.TrimRight(url, "/ ") + img + s
	//图片存储在本地文件系统中。
	case StoreLocal:
		url = img
	}
	fmt.Println(img)
	fmt.Println(url)
	return url
}

// Substr returns the substr from start to length.
func Substr(s string, length int) string {
	bt := []rune(s)
	start := 0
	dot := false

	if start > len(bt) {
		start = start % len(bt)
	}
	var end int
	if (start + length) > (len(bt) - 1) {
		end = len(bt)
	} else {
		end = start + length
		dot = true
	}

	str := string(bt[start:end])
	if dot {
		str = str + "..."
	}
	return str
}

// 判断数据是否在map中
func InMap(maps map[int]bool, key int) (ret bool) {
	if _, ok := maps[key]; ok {
		return true
	}
	return
}

// 分页函数（这个分页函数不具有通用性）
// rollPage:展示分页的个数
// totalRows：总记录
// currentPage:每页显示记录数
// urlPrefix:url链接前缀
// urlParams:url键值对参数
func NewPaginations(rollPage, totalRows, listRows, currentPage int, urlPrefix string, urlSuffix string, urlParams ...interface{}) html1.HTML {
	var (
		htmlPage, path string
		pages          []int
		params         []string
	)
	//总页数
	totalPage := totalRows / listRows
	if totalRows%listRows > 0 {
		totalPage += 1
	}
	//只有1页的时候，不分页
	if totalPage < 2 {
		return ""
	}
	paramsLen := len(urlParams)
	if paramsLen > 0 {
		if paramsLen%2 > 0 {
			paramsLen = paramsLen - 1
		}
		for i := 0; i < paramsLen; {
			key := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i]))
			val := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i+1]))
			//键存在，同时值不为0也不为空
			if len(key) > 0 && len(val) > 0 && val != "0" {
				params = append(params, key, val)
			}
			i = i + 2
		}
	}

	path = strings.Trim(urlPrefix, "/")
	if len(params) > 0 {
		path = path + "/" + strings.Trim(strings.Join(params, "/"), "/")
	}
	//最后再处理一次“/”，是为了防止urlPrifix参数为空时，出现多余的“/”
	path = "/" + strings.Trim(path, "/")

	if currentPage > totalPage {
		currentPage = totalPage
	}
	if currentPage < 1 {
		currentPage = 1
	}
	index := 0
	rp := rollPage * 2
	for i := rp; i > 0; i-- {
		p := currentPage + rollPage - i
		if p > 0 && p <= totalPage {

			pages = append(pages, p)
		}
	}
	for k, v := range pages {
		if v == currentPage {
			index = k
		}
	}
	pages_len := len(pages)
	if currentPage > 1 {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=1%v">1..</a></li><li><a class="num" href="`+path+`?page=%d%v">«</a></li>`, urlSuffix, currentPage-1, urlSuffix)
	}
	if pages_len <= rollPage {
		for _, v := range pages {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	} else {
		var pageSlice []int
		indexMin := index - rollPage/2
		indexMax := index + rollPage/2
		if indexMin > 0 && indexMax < pages_len { //切片索引未越界
			pageSlice = pages[indexMin:indexMax]
		} else {
			if indexMin < 0 {
				pageSlice = pages[0:rollPage]
			} else if indexMax > pages_len {
				pageSlice = pages[(pages_len - rollPage):pages_len]
			} else {
				pageSlice = pages[indexMin:indexMax]
			}

		}

		for _, v := range pageSlice {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	}
	if currentPage < totalPage {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%v%v">»</a></li><li><a class="num" href="`+path+`?page=%v%v">..%d</a></li>`, currentPage+1, urlSuffix, totalPage, urlSuffix, totalPage)
	}

	return html1.HTML(`<ul class="pagination">` + htmlPage + `</ul>`)
}
