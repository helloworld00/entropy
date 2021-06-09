package main

import(
	"path/filepath"
	"fmt"
	"flag"
	"os"
	"time"
)

type treeNode struct {
	size int64

	child *treeNode
	sibling *treeNode
	parent *treeNode

	isDir bool
	name string
	path string
}

func (node *treeNode) setInfo(info os.FileInfo, path string) {
	node.name = info.Name()
	node.isDir = info.IsDir()
	if !node.isDir {
		node.size = info.Size()
	}
	node.path = path
}

func (node *treeNode) output() {
	fmt.Println(node.path, node.name, node.size)
	if node.child != nil {
		node.child.output()
	}
	if node.sibling != nil {
		node.sibling.output()
	}
}

type result struct {
	rootNode *treeNode
	fileCount int64
	dirCount int64
	startTime time.Time
	rootPath string

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

func (ret *result) output() {
	node := ret.rootNode
	fmt.Println("")
	fmt.Println("File count: ", ret.fileCount, ", Dir count: ", ret.dirCount, "Duration: ", time.Now().Sub(ret.startTime))
	node.output()

	//duplicated
	fmt.Println("")
	fmt.Println("Possible duplicated:")
	for _,v := range ret.duplicatedMap {
		if len(v) < 2 {
			continue
		}
		//hashmap := make(map[string][]*treeNode)
		for _,n :=  range v {
			fmt.Println(n.path)
		}
		//hashmap = nil
		fmt.Println("")
	}
}

func main(){
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: ./entropy dirpath")
		os.Exit(0)
	}
	root := filepath.Clean(args[0])
	ret := walkThrough(root)
	ret.output()
	//wait debug
	//time.Sleep(time.Second * 5)
}

func walkThrough(root string) *result{
	ret := &result{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error{
		node := &treeNode{}
		node.setInfo(info, path)
		ret.insertNode(node)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return ret
}
