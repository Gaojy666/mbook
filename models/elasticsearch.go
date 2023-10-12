package models

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"strconv"
	"strings"
	"ziyoubiancheng/mbook/utils"
)

func ElasticBuildIndex(bookId int) {
	// 1.将book数据从mysql中取出，而后建立ES数据索引
	// 2.把这本书之下的所有章节拿出来，再建立ES章节索引

	// 根据book_id取出对应图书的name和description，连同id发送给elasticsearch建立索引
	book, _ := NewBook().Select("book_id", bookId, "book_id", "book_name", "description")
	addBookToIndex(book.BookId, book.BookName, book.Description)

	// index documents
	var documents []Document
	fields := []string{"document_id", "book_id", "document_name", "release"}
	//... 操作符进行展开传递,表示将切片中的元素逐个传递给 All 方法作为独立的参数。
	GetOrm("r").QueryTable(TNDocuments()).Filter("book_id", bookId).All(&documents, fields...)
	if len(documents) > 0 {
		for _, document := range documents {
			addDocumentToIndex(document.DocumentId, document.BookId, flatHtml(document.Release))
		}
	}
}

// 根据关键字按图书搜索，返回图书id数组,书的总数和err
func ElasticSearchBook(kw string, pageSize, page int) ([]int, int, error) {
	// 构建一个查询图书的json，发送给ES，从ES返回的结果解析出所有的图书id
	var ids []int
	count := 0

	if page > 0 {
		// ES的offset是从0开始的
		page = page - 1
	} else {
		page = 0
	}

	// multi_match是多字段查询
	// _source是过滤字段，表示返回的结果里用哪些字段
	queryJson := `
		{
			"query": {
				"multi_match": {
					"query": %v,
					"fields": ["book_name", "description"]
				}
			},
			"_source":["book_id"],
			"size": %v,
			"from": %v
		}	
	`

	// elasticsearch api
	host, _ := web.AppConfig.String("elastic_host")
	api := host + "mbooks/datas/_search"
	queryJson = fmt.Sprintf(queryJson, kw, pageSize, page)

	// 返回simpleJson对象，使用起来更方便
	sj, err := utils.HttpPostJson(api, queryJson)
	if err == nil {
		count = sj.GetPath("hits", "total").MustInt()
		resultArray := sj.GetPath("hits", "hits").MustArray()
		for _, v := range resultArray {
			if each_map, ok := v.(map[string]interface{}); ok {
				id, _ := strconv.Atoi(each_map["_id"].(string))
				ids = append(ids, id)
			}
		}
	}

	return ids, count, err
}

// 根据关键字按章节搜索，返回章节id数组,书的总数和err
func ElasticSearchDocument(kw string, pageSize, page int, bookId ...int) ([]int, int, error) {
	var ids []int
	count := 0

	if page > 0 {
		page = page - 1
	} else {
		page = 0
	}

	// 搜索全部图书
	queryJson := `
		{
			"query":{
				"match": {
					"release": "ajax",
				}
			},
			"_source":["document_id"],
			"size": %v,
			"from": %v
		}
	`
	queryJson = fmt.Sprintf(queryJson, kw, pageSize, page)

	// 按照图书搜索
	if len(bookId) > 0 && bookId[0] > 0 {
		queryJson = `
			{
				"query": {
					"bool": {
						"filter": [{
							"term": {
								"book_id":%v
							}
						}],
						"must": {
							"multi_match": {
								"query": "%v",
								"fields": ["release"]
							}
						}
					}
				},
				"from": %v,
				"size": %v,
				"_source": ["document_id"]
			}
		`
		queryJson = fmt.Sprintf(queryJson, kw, pageSize, page)
	}

	//elasticsearch api
	host, _ := web.AppConfig.String("elastic_host")
	api := host + "mdocuments/datas/_search"
	queryJson = fmt.Sprintf(queryJson, kw, pageSize, page)

	fmt.Println(api)
	fmt.Println(queryJson)

	sj, err := utils.HttpPostJson(api, queryJson)

	if err == nil {
		count = sj.GetPath("hits", "total").MustInt()
		resultArray := sj.GetPath("hits", "hits").MustArray()
		for _, v := range resultArray {
			if each_map, ok := v.(map[string]interface{}); ok {
				id, _ := strconv.Atoi(each_map["_id"].(string))
				ids = append(ids, id)
			}
		}
	}

	return ids, count, err
}

func addBookToIndex(bookId int, bookName string, description string) {
	// mbooks/datas/[bookid]
	queryJson := `
		{
			"book_id":%v,
			"book_name":"%v",
			"description":"%v"
		}
	`
	// elasticsearch api
	host, _ := web.AppConfig.String("elastic_host")
	api := host + "mbooks/datas/" + strconv.Itoa(bookId)

	// 发起请求
	queryJson = fmt.Sprintf(queryJson, bookId, bookName, description)
	err := utils.HttpPutJson(api, queryJson)
	if err != nil {
		logs.Debug(err)
	}
}

func addDocumentToIndex(documentId, bookId int, release string) {
	queryJson := `
		{
			"document_id":%v,
			"book_id":%v,
			"release":"%v"
		}
	`

	//elasticsearch api
	host, _ := web.AppConfig.String("elastic_host")
	api := host + "mdocuments/datas/" + strconv.Itoa(documentId)

	// 发起请求
	queryJson = fmt.Sprintf(queryJson, documentId, bookId, release)
	err := utils.HttpPutJson(api, queryJson)
	if err != nil {
		logs.Debug(err)
	}
}

// 对html字符串做预处理
func flatHtml(htmlStr string) string {
	// 1.将回车符剔除
	// 2.将双引号剔除

	htmlStr = strings.Replace(htmlStr, "\n", " ", -1)
	htmlStr = strings.Replace(htmlStr, "\"", "", -1)

	// goquery库解析html文本
	gq, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		// 无法解析
		return htmlStr
	}
	// 去掉前端标签内容，返回处理后的字符串
	return gq.Text()
}
