package main

import(
	"path/filepath"
	"fmt"
	"os"
	"time"
	"github.com/gin-gonic/gin"
)

var resultCache map[string]*result

func apiInit(){
	resultCache = make(map[string]*result)
}

func FailRequest(c *gin.Context, errcode int, errmsg string){
	if errcode == 0 {
		panic(1)
	}

	ret := BuildResponse(errcode, errmsg, nil)
	c.JSON(200, ret)
}

func BuildResponse(errcode int, errmsg string, result interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["errcode"] = errcode
	ret["errmsg"] = errmsg
	ret["result"] = result
	return ret
}

func FinishRequest(c *gin.Context, errmsg string, result interface{}){
	ret := BuildResponse(0, errmsg, result)
	c.JSON(200, ret)
}

func scan(c *gin.Context) {
	p := c.Query("path")
	p = filepath.Clean(p)
	if len(p) < 1 {
		FailRequest(c, -201, "path error")
		return
	}
	_, err := os.Stat(p)
	if err != nil {
		FailRequest(c, -201, "path not exists")
		return
	}

	res := resultCache[p]
	if res != nil {
		if !res.finished {
			FailRequest(c, -202, p + " in scanning. Try later.")
			return
		}
		res.cleanup()
	}
	
	resultCache[p] = new(result)
	go walkThrough(p)
	FinishRequest(c, "ok", nil)
}

func walkThrough(root string) {
	ret := resultCache[root]

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error{
		node := &treeNode{}
		node.setInfo(info, path)
		ret.insertNode(node)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		ret.errmsg = err.Error()
	}
	ret.finished = true
	ret.endTime = time.Now()
}

func getStatus(c *gin.Context) {
	p := c.Query("path")
	p = filepath.Clean(p)
	if len(p) < 1 {
		FailRequest(c, -201, "path error")
		return
	}
	res := resultCache[p]
	
	FinishRequest(c, "ok", res)
}

func getChildren(c *gin.Context) {
	p := c.Query("path")
	p = filepath.Clean(p)
	if len(p) < 1 {
		FailRequest(c, -201, "path error")
		return
	}
	res := resultCache[p]
	if res == nil {
		FailRequest(c, -201, "no info, scan first")
		return
	}

	nodepath := c.Query("nodepath")
	if len(nodepath) < 1 {
		nodepath = res.rootPath
	}
	
	FinishRequest(c, "ok", res.getChildren(nodepath))
}

func remove(c *gin.Context) {

}

func getDuplicated(c *gin.Context) {

}
