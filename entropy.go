package main

import(
	"path/filepath"
	"fmt"
	"os"
	"time"
	"runtime"
	"sort"
	"encoding/json"
)

type treeNode struct {
	size int64

	child *treeNode
	sibling *treeNode
	parent *treeNode

	isDir bool
	name string
	path string
	modified time.Time
}

func (node *treeNode) MarshalJSON() ([]byte, error){
	retdic := make(map[string]interface{})
	retdic["size"] = node.size
	retdic["sizeStr"] = node.sizeStr()
	retdic["isDir"] = node.isDir
	retdic["name"] = node.name
	retdic["path"] = node.path
	retdic["modified"] = node.modified.Format("2006-01-02 15:04:05")
	return json.Marshal(retdic)
}

func (node *treeNode) setInfo(info os.FileInfo, path string) {
	node.name = info.Name()
	node.isDir = info.IsDir()
	if !node.isDir {
		node.size = info.Size()
	}
	node.path = path
	node.modified = info.ModTime()
}

func (node *treeNode) sizeStr() string {
	size := node.size
	if size < 1024 {
	   return fmt.Sprintf("%.0fB", float64(size) )
	} else if size < (1024 * 1024) {
	   return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
	} else if size < (1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fMB", float64(size)/float64(1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fGB", float64(size)/float64(1024*1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024 * 1024) {
	   return fmt.Sprintf("%.2fTB", float64(size)/float64(1024*1024*1024*1024))
	} else {
	   return fmt.Sprintf("%.2fEB", float64(size)/float64(1024*1024*1024*1024*1024))
	}
}

type result struct {
	rootNode *treeNode
	fileCount int64
	dirCount int64
	startTime time.Time
	endTime time.Time
	rootPath string

	finished bool
	errmsg string

	nodeCache map[string]*treeNode
	duplicatedMap map[int64][]*treeNode
}

func (ret *result) insertNode(node *treeNode) {
	if ret.rootNode == nil {
		ret.rootNode = node
		ret.startTime = time.Now()
		ret.rootPath = node.path
		ret.nodeCache = make(map[string]*treeNode)
		ret.nodeCache[node.path] = node
		ret.duplicatedMap = make(map[int64][]*treeNode)
		ret.duplicatedMap[node.size] = append(ret.duplicatedMap[node.size], node)
		return
	}
	dir := filepath.Dir(node.path)
	parentNode := ret.nodeCache[dir]
	node.parent = parentNode

	if parentNode.child == nil {
		parentNode.child = node
	} else {
		pre := parentNode.child
		for pre.sibling != nil {
			pre = pre.sibling
		}
		pre.sibling = node
	}
	//update parent size
	if !node.isDir {
		for parentNode != nil {
			parentNode.size += node.size
			parentNode = parentNode.parent
		}
	}

	if node.isDir {
		ret.dirCount++
	} else {
		ret.fileCount++
	}
	ret.nodeCache[node.path] = node
	ret.duplicatedMap[node.size] = append(ret.duplicatedMap[node.size], node)
}

func (ret *result) getChildren(p string) []*treeNode{
	parent := ret.nodeCache[p]
	if parent == nil || parent.child == nil {
		return nil
	}
	children := make([]*treeNode, 0, 10)
	child := parent.child
	for child != nil {
		children = append(children, child)
		child = child.sibling
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].size > children[j].size
	})
	return children
}

func (ret *result) MarshalJSON() ([]byte, error){
	retdic := make(map[string]interface{})
	retdic["fileCount"] = ret.fileCount
	retdic["dirCount"] = ret.dirCount
	if ret.finished {
		retdic["duration"] = fmtDuration(ret.endTime.Sub(ret.startTime))
	} else {
		retdic["duration"] = fmtDuration(time.Now().Sub(ret.startTime))
	}
	retdic["rootPath"] = ret.rootPath
	retdic["finished"] = ret.finished
	retdic["errmsg"] = ret.errmsg
	retdic["totalSize"] = ""
	if ret.rootNode != nil {
		retdic["totalSize"] = ret.rootNode.sizeStr()
	}
	return json.Marshal(retdic)
}

func fmtDuration(d time.Duration) string {
    d = d.Round(time.Second)
    h := d / time.Hour
    d -= h * time.Hour
    m := d / time.Minute
    d -= m * time.Minute
	s := d / time.Second
    return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (ret *result) cleanup() {
	ret.rootNode = nil
	ret.nodeCache = nil
	ret.duplicatedMap = nil
	runtime.GC()
}
