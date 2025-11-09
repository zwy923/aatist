package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// 定义事件结构体
type Event struct {
	Service   string `json:"service"`   // 来源服务 (e.g., "AI", "Backend", "Scraper")
	Type      string `json:"type"`      // 事件类型 (e.g., "INFO", "ERROR", "ACTION")
	Message   string `json:"message"`   // 事件内容
	Timestamp string `json:"timestamp"` // 时间戳
}

// 日志文件路径
const logFilePath = "./logs/events.log"

// 初始化日志文件
func initLogFile() {
	os.MkdirAll("./logs", os.ModePerm)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("❌ 无法打开日志文件: %v", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 写入事件到文件
func logEvent(event Event) {
	event.Timestamp = time.Now().Format(time.RFC3339)
	data, _ := json.Marshal(event)
	log.Println(string(data))
}

// Gin 路由
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/log", func(c *gin.Context) {
		var event Event
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		logEvent(event)
		c.JSON(http.StatusOK, gin.H{"message": "event logged"})
	})

	return r
}

func main() {
	initLogFile()
	r := setupRouter()

	port := "8081"
	fmt.Printf("🚀 Event Logger Service running on http://localhost:%s\n", port)
	r.Run(":" + port)
}
