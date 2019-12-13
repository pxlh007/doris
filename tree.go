package doris

import (
	"strings"
)

type (
	// 路由结构的定义
	tree struct {
		root   *node  // 树的根节点
		method string // HTTP方法
		doris  *Doris // 对应框架实例的指针
	}
	// 以method作为Key
	// 所有的路由按照method
	// 分成多颗树查找的时候根据method
	// 定位到对应的树然后再根据path检索树
	// 找到对应的函数链之后逐个执行
	trees map[string]*tree // 保存全部的HTTP方法路由树

	// 对应节点结构
	node struct {
		nType    nodeType      // 节点类型：普通，参数，全匹配
		label    byte          // 节点检索首字母
		prefix   string        // 节点前缀
		priority uint32        // 节点优先级（用于冲突时候选择）
		parent   *node         // 指向父节点
		children children      // 指向子节点
		fullPath string        // 路由全路径
		pList    Params        // 参数列表
		handlers HandlersChain // 函数处理链
		debug    bool          // debug开关
	}
	// 保存节点值结构
	nodeValue struct {
		handlers HandlersChain
		params   Params
		pvalues  []interface{}
		fullPath string
	}
	nodeType uint8    // 节点类型
	children []*node  // 节点切片
	Params   []string // 参数列表
)

// 路由类型常量
const (
	skind nodeType = iota // 常数（默认）
	pkind                 // 参数
	akind                 // 全匹配
)

// 创建新路由树
func newTree(doris *Doris, method string) *tree {
	return &tree{
		root:   doris.trees[method].root,
		method: method,
		doris:  doris,
	}
}

// 创建新节点
func newNode(
	t nodeType,
	pre string,
	p *node,
	c []*node,
	fp string,
	pl Params,
	h HandlersChain) *node {
	return &node{
		nType:    t,
		label:    pre[0],
		prefix:   pre,
		parent:   p,
		children: c,
		fullPath: fp,
		pList:    pl,
		handlers: h,
	}
}

// 创建节点值
func newNodeValue(
	handlers HandlersChain,
	params Params,
	pv []interface{},
	fp string) *nodeValue {
	return &nodeValue{
		handlers: handlers,
		params:   params,
		pvalues:  pv,
		fullPath: fp,
	}
}

// 流程控制
func (n *node) addRoute(path string, handlers HandlersChain) {
	// 处理path
	if len(path) == 0 {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}
	// 参数初始化
	var (
		pList                        Params
		tmpPList                     Params // 存储临时参数列表
		tmpHandlers                  HandlersChain
		param, fullPath              string
		i, j, lp, lc, counter, index int
		cn, child                    *node
	)
	// 原始路径
	fullPath = path
	cn = n
	// 循环path
	for path != "" {
		i = 0
		lp = len(path)
		lc = len(cn.prefix)
		tmpPList = Params{}
		// 计算循环参数
		counter = lp
		if counter > lc {
			counter = lc
		}
		// 相等继续循环
		for ; i < counter && path[i] == cn.prefix[i]; i++ {
			continue
		}

		// /:sex
		debugPrintMessage("i========", i, n.debug)  // 1
		debugPrintMessage("lc=======", lc, n.debug) // 5
		debugPrintMessage("lp=======", lp, n.debug) // 5

		// 未达尾部
		if i < lc { // 超过path长度
			if i >= lp {
				cn.nodeFission(i)
				// 更新当前节点参数
				cn.prefix = path
				cn.handlers = handlers
				break
			}
			// 出现不等情况
			if path[i] == ':' { // 参数路由
				debugPrintMessage("进入参数路由区域", "__print__", n.debug)
				debugPrintMessage("更新之前的cn", "__print__", n.debug)
				debugPrintMessage("cn", cn, n.debug)
				debugPrintMessage("path", path, n.debug)
				debugPrintMessage("cn.parent", cn.parent, n.debug)
				j = 0
				for j = i + 1; j < lp && path[j] != '/'; j++ {
					continue
				}
				// 裂变节点
				debugPrintMessage("裂变前=====", "__print__", n.debug)
				debugPrintMessage("cn", cn, n.debug)
				cn.nodeFission(i)

				debugPrintMessage("裂变后=====", "__print__", n.debug)
				debugPrintMessage("cn====", cn, n.debug)

				// 提取参数
				param = path[i+1 : j]
				path = path[j:]
				fullPath = cn.fullPath + ":" + param
				// 更新参数
				pList = append(pList, param)
				// path到尾部
				if path == "" {
					tmpHandlers = handlers
					tmpPList = pList
				}
				// 给cn插入新节点并更新cn
				// 先查找是否存在参数:节点
				pChild := cn.findChildByKind(pkind)
				if pChild == nil {
					// debugPrintMessage("len(pChild.pList)", len(pChild.pList), n.debug)
					// debugPrintMessage("pChild.pList", pChild.pList, n.debug)
					pChild = cn.insertNode(pkind, ":", fullPath, children{}, tmpPList, tmpHandlers)
				} else if path == "" {
					// 使用新参数覆盖旧的
					pChild.pList = tmpPList
					pChild.handlers = handlers
				}
				// 更新cn
				cn = pChild
				debugPrintMessage("更新之后的cn", "__print__", n.debug)
				debugPrintMessage("cn", cn, n.debug)
			} else if path[i] == '*' { // 全路由
				// 裂变节点
				cn.nodeFission(i)
				// 插入全匹配节点
				pList = append(pList, "*")
				fullPath = cn.fullPath + "*"
				cn = cn.insertNode(akind, "*", fullPath, children{}, pList, handlers)
				// 碰到*说明到达末尾跳出循环
				break
			} else { // 静态路由
				debugPrintMessage("进入静态存储区", "__print__", n.debug)
				debugPrintMessage("i", i, n.debug)
				debugPrintMessage("path", path, n.debug)
				// 裂变节点
				cn.nodeFission(i)
				path = path[i:]
				indexP := strings.Index(path, ":")
				indexA := strings.Index(path, "*")
				if indexP == -1 && indexA == -1 { // 说明剩余的部分既不含:也不含*
					fullPath = cn.fullPath + path
					// 插入静态节点
					cn = cn.insertNode(skind, path, fullPath, children{}, Params{}, handlers)
					// path已经处理完
					break
				} else { // 含:或者*的提取:或者*之前的部分
					// 查找子节点
					debugPrintMessage("断点4 path", path, n.debug)
					debugPrintMessage("断点4 i", i, n.debug)
					child = cn.findChildByLabel(path)
					debugPrintMessage("断点4 child", child, n.debug)
					if child != nil {
						// 含有子节点
						cn = child
					} else {
						if indexP > indexA {
							index = indexP
						} else {
							index = indexA
						}
						prefix := path[:index]
						fullPath = cn.fullPath + path[:index]
						path = path[index:]
						// 插入静态节点
						cn = cn.insertNode(skind, prefix, fullPath, children{}, Params{}, HandlersChain{})
					}
					debugPrintMessage("cn", cn, n.debug)
				}
			}
		} else { // 到达当前节点尾部
			debugPrintMessage("到达当前节点尾部", "__print__", n.debug)
			fullPath := cn.fullPath + path[i:]
			prefix := path[i:]
			path = path[i:]
			// debugPrintMessage("断点3 cn.prefix", cn.prefix, n.debug)
			// debugPrintMessage("断点3 cn.fullPath", cn.fullPath, n.debug)
			// debugPrintMessage("断点3 path", path, n.debug)
			// debugPrintMessage("断点3 prefix", prefix, n.debug)
			// 重复路由处理
			if path == "" {
				if len(cn.handlers) == 0 {
					cn.handlers = handlers
				}
				break
			}
			// 无子节点且当前节点无处理函数将剩余部分粘到cn
			// len(cn.handlers) == 0 解决只有一个分支的时候下面是参数节点
			// 上面节点合并/导致路由错误的问题。
			if len(cn.children) == 0 && len(cn.handlers) == 0 {
				indexP := strings.Index(path, ":")
				indexA := strings.Index(path, "*")
				if indexP == -1 && indexA == -1 {
					if len(cn.handlers) == 0 {
						cn.prefix += path
						cn.fullPath += path
						if len(handlers) != 0 {
							cn.handlers = handlers
						}
						debugPrintMessage("当前节点handlers", "__print__", n.debug)
						debugPrintMessage("cn.handlers", cn.handlers, n.debug)
					} else { // 插入新节点
						debugPrintMessage("path", path, n.debug)
						debugPrintMessage("cn.fullPath", cn.fullPath, n.debug)
						_ = cn.insertNode(skind, prefix, fullPath, children{}, Params{}, handlers)
					}
					break
				} else if indexP > indexA { // 包含:
					index = indexP
				} else { // 包含*
					index = indexA
				}
				// 更新当前节点的参数
				// 更新path参数
				cn.prefix += path[:index]
				cn.fullPath += path[:index]
				path = path[index:]
			} else { // 存在子节点
				child = cn.findChildByLabel(path)
				if child != nil {
					if path[0] == ':' {
						continue
					} else if path[0] == '*' {
						break
					} else { // 更新参数继续循环
						cn = child
					}
				} else { // 未找到子节点
					if path[0] == ':' || path[0] == '*' {
						continue
					} else {
						prefix := string(path[0])
						fullPath = cn.fullPath + string(path[0])
						tmpHandlers := HandlersChain{}
						// 需注释否则有部分参数丢失
						// pList = Params{}
						child = cn.insertNode(skind, prefix, fullPath, children{}, tmpPList, tmpHandlers)
						cn = child
					}
				}
			}
		}
	}
}

// 裂变节点流程
// 职责：根据传递进来的裂变位置将当前节点
// 分割为两个独立的节点，并更新参数
func (n *node) nodeFission(i int) {
	// debugPrintMessage("nodeFission i===", i, true)
	if i == 0 {
		return
	}
	// debugPrintMessage("nodeFission i===", i, true)
	// debugPrintMessage("nodeFission n===", n, true)
	// 设置参数
	prefix := n.prefix
	fullPath := n.fullPath
	pList := n.pList
	handlers := n.handlers
	n.prefix = prefix[:i]
	newChildren := n.children
	// debugPrintMessage("nodeFission n===", n, true)
	// debugPrintMessage("nodeFission n.prefix ===", prefix, true)
	// debugPrintMessage("nodeFission n.fullPath===", fullPath, true)
	// n.fullPath = fullPath[:i]
	n.fullPath = fullPath[:len(fullPath)-len(prefix[i:])] // 调整计算全路径方式
	n.handlers = []HandlerFunc{}
	n.pList = Params{}
	n.children = children{}
	// 插入节点
	_ = n.insertNode(skind, prefix[i:], fullPath, newChildren, pList, handlers)
}

// 插入节点流程
// 职责：根据传递进来的参数在当前节点的尾部插入指定类型的节点
func (n *node) insertNode(
	kind nodeType,
	prefix,
	fullPath string,
	children children,
	pList Params,
	handlers HandlersChain) *node {
	// 创建新节点
	node := newNode(kind, prefix, n, children, fullPath, pList, handlers)
	// 添加子节点
	n.addChild(node)
	// 返回新节点
	return node
}

// 添加子节点
func (n *node) addChild(nn *node) {
	// 插入子节点
	n.children = append(n.children, nn)
}

// 根据label和kind查找子节点
func (n *node) findChild(label string, kind nodeType) *node {
	ln := len(n.children)
	//	debugPrintMessage("findChild ln", ln, true)
	for i := 0; i < ln; i++ {
		// 循环所有的子节点
		if n.children[i].label == label[0] && n.children[i].nType == kind {
			// 找到包含path首字母的子节点
			return n.children[i]
		}
	}
	return nil
}

// 根据label查找子节点
func (n *node) findChildByLabel(label string) *node {
	ln := len(n.children)
	for i := 0; i < ln; i++ {
		// 循环所有的子节点
		if n.children[i].label == label[0] {
			// 找到包含path首字母的子节点
			return n.children[i]
		}
	}
	return nil
}

// 使用kind查询子节点
func (n *node) findChildByKind(kind nodeType) *node {
	ln := len(n.children)
	for i := 0; i < ln; i++ {
		// 循环所有的子节点
		if n.children[i].nType == kind {
			// 找到包含path首字母的子节点
			return n.children[i]
		}
	}
	return nil
}

// 使用kind查找当前节点全部子节点
func (n *node) findChildrenByKind(kind nodeType) []*node {
	ln := len(n.children)
	cns := []*node{}
	for i := 0; i < ln; i++ {
		// 循环所有的子节点
		if n.children[i].nType == kind {
			// 找到包含path首字母的子节点
			cns = append(cns, n.children[i])
		}
	}
	return cns
}

// 查找路由
func (n *node) find(path string) (nv *nodeValue) {
	// 查找具体路由
	// 查找优先级：静态 > 参数 > 全量
	// 初始化参数
	var (
		search            = path
		cn                = n
		child             *node         // 子节点
		nn                *node         // 设定下一个搜索节点
		nk                nodeType      // 设定下一个搜索类型
		ns                string        // 设定下一个搜索字串
		pvalues           []interface{} // 保存参数值
		max, l, lp, i, ls int           // 计数器参数
		count             int           // 参数计数器
	)
	// 进入大循环
	for {
		// 优先处理static路由
		if search == "" || cn == nil {
			break
		}
		// debug
		debugPrintMessage("循环开始位置", "__print__", n.debug)
		debugPrintMessage("cn", cn, n.debug)
		debugPrintMessage("cn.label", cn.label, n.debug)
		debugPrintMessage("cn.prefix", cn.prefix, n.debug)
		debugPrintMessage("search", search, n.debug)
		// 初始化max参数
		// 取ls和lp中的小者
		ls = len(search)
		lp = len(cn.prefix)
		max = lp
		if max > ls {
			max = ls
		}
		debugPrintMessage("max", max, n.debug)
		// 计算参数字符
		l = 0
		if cn.label != ':' {
			// 循环比对前缀
			for ; l < max && cn.prefix[l] == search[l]; l++ {
				// 跳过相同的部分
			}
		}
		debugPrintMessage("lp", lp, n.debug)
		debugPrintMessage("l", l, n.debug)
		if l == lp {
			// 已经到达前缀结束
			search = search[l:] // 更新search
			debugPrintMessage("到达当前节点前缀末尾", "__print__", n.debug)
			debugPrintMessage("search", search, n.debug)
		} else {
			// 前缀没有匹配到
			if nn == nil {
				return // 404错误
			} else {
				// 分三种情况讨论
				// 情况1：search用完
				// 情况2：当前节点为参数节点
				// 情况3：普通情况search未用完
				// 仅情况2需要先查静态子节点
				if cn.label != ':' {
					// 返回上一级查参数子节点
					cn = nn
					search = ns
					if nk == pkind {
						goto paramProcess
					} else if nk == akind {
						goto allProcess
					}
				}
			}
		}
		debugPrintMessage("search", search, cn.debug)
		// 检查新search
		if search == "" {
			break
		}
		// 查静态节点
		if child = cn.findChild(search, skind); child != nil {
			debugPrintMessage("child", child, cn.debug)
			debugPrintMessage("search", search, cn.debug)
			// 找到子节点
			// 当前节点前缀中末尾是/
			// 认为要么是参数路由要么是全匹配路由
			if cn.prefix[len(cn.prefix)-1] == '/' { // 解决静态路由和动态路由冲突问题
				nk = pkind  // 下一个节点是参数节点
				nn = cn     // 下一个节点设置为当前节点
				ns = search // 下一个节点的搜索关键词
			}
			cn = child
			continue
		}
		// 参数节点
	paramProcess:
		// param路由处理子流程
		debugPrintMessage("参数处理子流程", "__print__", n.debug)
		if child = cn.findChildByKind(pkind); child != nil {
			// Issue #378 Fix routing with slash included in parameter value
			// 这个地方待验证
			if len(pvalues) == count {
				// 验证完毕开启
				// continue
			}
			// 当前节点前缀中末尾是/
			// 认为要么是参数路由要么是全匹配路由
			if cn.prefix[len(cn.prefix)-1] == '/' { // 解决静态路由和动态路由冲突问题
				nk = akind  // 下一个节点是全匹配节点?
				nn = cn     // 下一个节点设置为当前节点
				ns = search // 下一个节点的搜索关键词
			}
			// 获取参数值
			l = len(search)
			i = 0 // 重新初始化i防止溢出
			//			debugPrintMessage("----i", i, n.debug)
			for ; i < l && search[i] != '/'; i++ {
				// 跳过中间部分
			}
			debugPrintMessage("cn", cn, n.debug)
			debugPrintMessage("child", child, n.debug)
			debugPrintMessage("search", search, n.debug)
			debugPrintMessage("pvalues", pvalues, n.debug)
			pvalues = append(pvalues, search[:i])
			search = search[i:]
			cn = child
			continue
		}
		// 全匹配节点
	allProcess:
		// all路由处理子流程
		// 全匹配节点的查找放到最后
		debugPrintMessage("全匹配处理子流程", "__print__", n.debug)
		if cn = cn.findChildByKind(akind); cn == nil {
			// 找到子节点
			if nn != nil {
				cn = nn
				nn = cn.parent // 下一个查找节点
				if nn != nil {
					// nk = nn.nType
				}
				search = ns
				if nk == pkind {
					goto paramProcess
				} else if nk == akind {
					goto allProcess
				}
			}
			return // Not found
		}
		// 添加参数值
		pvalues = append(pvalues, search)
		break
	}
	// 判断返回
	debugPrintMessage("cn", cn, n.debug)
	if cn != nil && len(cn.handlers) != 0 {
		debugPrintMessage("开始组织结果", "__print__", n.debug)
		// 组装nodeValue值
		nv = &nodeValue{
			handlers: cn.handlers,
			params:   cn.pList,
			pvalues:  pvalues,
			fullPath: cn.fullPath,
		}
	}
	// 路由结束
	debugPrintMessage("路由查找完毕", "__print__", n.debug)
	return
}

// 查找方法树
func (ts trees) get(method string) *node {
	if t, ok := ts[method]; ok {
		return t.root
	}
	return nil
}
