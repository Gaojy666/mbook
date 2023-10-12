package models

import (
	"bytes"
	"github.com/beego/beego/v2/server/web"
	"html/template"
	"strconv"
)

// 读书编辑页，图书目录
type DocumentMenu struct {
	DocumentId   int             `json:"id"`
	DocumentName string          `json:"text"`
	BookIdentify string          `json:"-"`
	Identify     string          `json:"identify"`
	ParentId     interface{}     `json:"parent"`
	Version      int64           `json:"version"`
	State        *highlightState `json:"state,omitempty"` //如果字段为空，则json中不会有该字段
}
type highlightState struct {
	Selected bool `json:"selected"`
	Opened   bool `json:"opened"`
}

// 生成一个表示菜单的 HTML 字符串
func (m *Document) GetMenuHtml(bookId, selectedId int) (string, error) {
	//通过调用 m.GetMenu(bookId, selectedId) 方法获取菜单树结构，
	//并将结果存储在 trees 变量中
	trees, err := m.GetMenu(bookId, selectedId)

	if err != nil {
		return "", err
	}

	// 根据菜单目录，找到当前选中节点的父节点
	parentId := m.highlightNode(trees, selectedId)

	//可变大小的缓冲区，可以用于高效地构建字符串，而不需要频繁地重新分配内存。
	buf := bytes.NewBufferString("")

	//根据菜单树结构生成 HTML 字符串，并将结果写入到 buf 中。
	// parent一开始传0,是因为书的第一章,肯定没有父节点
	m.treeHTML(trees, 0, selectedId, parentId, buf)

	return buf.String(), nil
}

func (m *Document) GetMenu(bookId int, selectedId int, isEdit ...bool) ([]*DocumentMenu, error) {
	//创建一个空的 []*DocumentMenu 类型的切片 trees，用于存储菜单数据。
	trees := make([]*DocumentMenu, 0)
	//创建一个 []*Document 类型的切片 docs，用于存储查询结果的某些字段。
	var docs []*Document

	// 查询与该书有关的章节数，并按照order_sort和identify排列
	// 并将查询结果的某些字段放入docs中
	count, err := GetOrm("r").QueryTable(m).Filter("book_id", bookId).OrderBy("order_sort", "identify").Limit(2000).All(&docs, "document_id", "document_name", "parent_id", "identify", "version")
	if err != nil {
		return trees, err
	}

	// 根据bookid查询相关书的信息
	book, _ := NewBook().Select("book_id", bookId)

	trees = make([]*DocumentMenu, count)
	for index, item := range docs {
		tree := &DocumentMenu{}

		if selectedId > 0 {
			//如果当前元素的 DocumentId 等于 selectedId，
			//将 tree 的 State 字段设置为一个具有选中和展开状态的 highlightState 对象。
			if selectedId == item.DocumentId {
				tree.State = &highlightState{
					Selected: true,
					Opened:   true,
				}
			}
		} else {
			//如果是第一章，
			//将 tree 的 State 字段设置为一个具有选中和展开状态的对象。
			if index == 0 {
				tree.State = &highlightState{
					Selected: true,
					Opened:   true,
				}
			}
		}

		//将当前元素的相关信息赋值给 tree 对象的相应字段。
		tree.DocumentId = item.DocumentId
		tree.Identify = item.Identify
		tree.Version = item.Version
		tree.BookIdentify = book.Identify

		// 父节点存在
		if item.ParentId > 0 {
			tree.ParentId = item.ParentId
		} else {
			// 该节点为根节点
			tree.ParentId = "#"
		}
		idf := item.Identify
		if idf == "" {
			idf = strconv.Itoa(item.DocumentId)
		}

		// 如果要编辑该章
		if len(isEdit) > 0 && isEdit[0] == true {
			// DocumentName字段在前端显示时会展示不同的颜色
			tree.DocumentName = item.DocumentName + "<small class='text-danger'>(" + idf + ")</small>"
		} else {
			tree.DocumentName = item.DocumentName
		}

		trees[index] = tree
	}

	return trees, nil
}

// 在给定的节点数组中查找具有特定父节点 ID 的节点，并返回匹配节点的 DocumentId
func (m *Document) highlightNode(array []*DocumentMenu, parentId int) int {
	for _, item := range array {
		// 判断 item.ParentId 是否为字符串类型
		if _, ok := item.ParentId.(string); ok && item.DocumentId == parentId {
			//该节点是根节点，且不具有有效的父节点

			//返回该节点的 DocumentId
			return item.DocumentId
		} else if pid, ok := item.ParentId.(int); ok && item.DocumentId == parentId {
			//该节点具有有效的父节点

			//如果 item.DocumentId 等于 parentId，则判断 item.ParentId 是否等于 parentId。
			//如果相等，说明该节点是自己的父节点，
			//即该节点是根节点。在这种情况下，返回 0 表示没有有效的父节点。
			if pid == parentId {
				return 0
			}
			// 递归寻找父节点
			return m.highlightNode(array, pid)
		}
	}
	return 0
}

// 生成目录的HTML结构
func (m *Document) treeHTML(array []*DocumentMenu, parentId int, selectedId int, selectedParentId int, buf *bytes.Buffer) {
	// 传入的buf为空
	// 首先向 buf 中写入 <ul> 标签，表示一个无序列表的开始
	buf.WriteString("<ul>")

	for _, item := range array {
		pid := 0
		//将 item.ParentId 转换为整数类型，并将结果赋值给 pid
		if p, ok := item.ParentId.(int); ok {
			pid = p
		}

		//如果 pid 等于 parentId，说明当前节点item是给定 parentId 的子节点，
		//需要包含在生成的目录 HTML 中。
		if pid == parentId {
			//根据节点的选中状态，设置 selected 和 selectedLi 字符串，用于指定节点的选中样式。
			selected := ""
			if item.DocumentId == selectedId {
				selected = ` class="jstree-clicked"`
			}
			selectedLi := ""
			if item.DocumentId == selectedParentId {
				selectedLi = ` class="jstree-open"`
			}
			//向 buf 中写入 <li> 标签，表示一个列表项的开始
			buf.WriteString("<li id=\"")
			buf.WriteString(strconv.Itoa(item.DocumentId))
			buf.WriteString("\"")
			buf.WriteString(selectedLi)
			buf.WriteString("><a href=\"")
			//根据节点的类型（item.Identify 是否为空），构建节点链接的 URL，并将其写入到 buf 中。
			if item.Identify != "" {
				uri := web.URLFor("DocumentController.Read", ":key", item.BookIdentify, ":id", item.Identify)
				buf.WriteString(uri)
			} else {
				uri := web.URLFor("DocumentController.Read", ":key", item.BookIdentify, ":id", item.DocumentId)
				buf.WriteString(uri)
			}
			//设置节点的标题属性（title）为节点的文档名称，并对其进行 HTML 转义
			buf.WriteString("\" title=\"")
			buf.WriteString(template.HTMLEscapeString(item.DocumentName) + "\"")
			buf.WriteString(selected + ">")
			buf.WriteString(template.HTMLEscapeString(item.DocumentName) + "</a>")

			//遍历 array 切片查找当前节点的子节点，如果找到子节点，则递归调用 treeHTML 方法处理子节点。
			for _, sub := range array {
				if p, ok := sub.ParentId.(int); ok && p == item.DocumentId {
					// sub是当前节点item的子节点，也就是item存在子节点
					m.treeHTML(array, p, selectedId, selectedParentId, buf)
					break
				}
			}
			//向 buf 中写入 </li> 标签，表示一个列表项的结束
			buf.WriteString("</li>")
		}
	}
	//向 buf 中写入 </ul> 标签，表示无序列表的结束。
	buf.WriteString("</ul>")
}
