package main

import(
	"github.com/gin-gonic/gin"
	"os/exec"
	"time"
	"fmt"
)

func main(){
	apiInit()

	server := gin.Default()

	server.Static("/static", "./static")

	gapi := server.Group("/api")
	{
		gapi.GET("/scan", scan)
		gapi.GET("/status", getStatus)
		gapi.GET("/children", getChildren)

		gapi.GET("/remove", remove)
		gapi.GET("/duplicated", getDuplicated)
	}
	go func(){
		time.Sleep(1 * time.Second)
		openPage()
	}()
	server.Run(":65375")
}

func openPage() {
	url := "http://localhost:65375/static/"
	fmt.Println("------------------------------------------------------------")
	fmt.Println("Please visit", url, "in browser")
	fmt.Println("------------------------------------------------------------")

	_, err := exec.Command("which", "open").Output()
	if err != nil {
		return
	}

	exec.Command("open", url).Output()
}

