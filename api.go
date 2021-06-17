package main

import(
	"path/filepath"
	"fmt"
	"os"
	"time"
	"github.com/gin-gonic/gin"
	"strings"
	"strconv"
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
		if info == nil || info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		node := &treeNode{}
		node.setInfo(info, path)
		ret.insertNode(node)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		ret.errmsg = err.Error()
	}
	ret.genDuplicatedKeys()
	ret.finished = true
	ret.endTime = time.Now()
}

func getStatus(c *gin.Context) {
	res := getResFromRequest(c)
	FinishRequest(c, "ok", res)
}

func getChildren(c *gin.Context) {
	res := getResFromRequest(c)
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
	res := getResFromRequest(c)
	if res == nil {
		FailRequest(c, -201, "no info, scan first")
		return 
	}

	nodepath := c.Query("nodepath")
	if len(nodepath) < 1 || res.nodeCache[nodepath] == nil {
		FailRequest(c, -201, "path error")
		return
	}

	err := os.RemoveAll(nodepath)
	if err != nil {
		FailRequest(c, -201, err.Error() )
		return
	}
	node := res.nodeCache[nodepath]
	res.removeNode(node)
	reloadPath, _ := filepath.Rel(res.rootPath, filepath.Dir(nodepath))
	FinishRequest(c, "ok", strings.Split(reloadPath, string(os.PathSeparator)) )
}

func getResFromRequest(c *gin.Context) *result {
	p := c.Query("path")
	p = filepath.Clean(p)
	if len(p) < 1 {
		return nil
	}
	res := resultCache[p]
	if res == nil {
		return nil
	}
	return res
}

func getDuplicated(c *gin.Context) {
	res := getResFromRequest(c)
	if res == nil {
		FailRequest(c, -201, "no info, scan first")
		return 
	}
	pno := c.DefaultQuery("pageno", "1")
	psz := c.DefaultQuery("pagesize", "20")
	pnoi, err1 := strconv.Atoi(pno)
	pszi, err2 := strconv.Atoi(psz)
	if err1 != nil {
		pnoi = 1
	}
	if err2 != nil {
		pszi = 20
	}
	data, finished := res.getDuplicated(pnoi, pszi)
	ret := make(map[string]interface{})
	ret["data"] = data
	ret["finished"] = finished
	FinishRequest(c, "ok",  ret)
}
