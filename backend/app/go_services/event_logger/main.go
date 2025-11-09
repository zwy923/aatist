package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 定义事件结构体
type Event struct {
	Service   string `json:"service"`            // 来源服务 (e.g., "AI", "Backend", "Scraper")
	Type      string `json:"type"`               // 事件类型 (e.g., "INFO", "ERROR", "ACTION")
	Message   string `json:"message"`            // 事件内容
	Timestamp string `json:"timestamp"`          // 时间戳
	Metadata  string `json:"metadata,omitempty"` // 额外元数据
}

// 日志文件路径
const logFilePath = "./logs/events.log"
const streamName = "event_logs"
const consumerGroup = "event_logger_group"
const consumerName = "event_logger_1"

var redisClient *redis.Client
var ctx = context.Background()

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

// 初始化 Redis 连接
func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("❌ Redis URL 解析失败: %v", err)
	}

	redisClient = redis.NewClient(opt)

	// 测试连接
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}

	// 创建消费者组（如果不存在）
	err = redisClient.XGroupCreateMkStream(ctx, streamName, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("⚠️ 创建消费者组失败: %v", err)
	}

	fmt.Println("✅ Redis 连接成功")
}

// 写入事件到文件
func logEvent(event Event) {
	event.Timestamp = time.Now().Format(time.RFC3339)
	data, _ := json.Marshal(event)
	log.Println(string(data))
}

// 从 Redis Stream 消费消息
func consumeFromStream() {
	for {
		// 读取消息
		streams, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    time.Second * 5,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue // 没有消息，继续等待
			}
			log.Printf("⚠️ 读取 Stream 失败: %v", err)
			time.Sleep(time.Second * 2)
			continue
		}

		// 处理消息
		for _, stream := range streams {
			for _, message := range stream.Messages {
				event := Event{
					Service: message.Values["service"].(string),
					Type:    message.Values["type"].(string),
					Message: message.Values["message"].(string),
				}
				if metadata, ok := message.Values["metadata"].(string); ok {
					event.Metadata = metadata
				}

				// 写入日志文件
				logEvent(event)

				// 确认消息已处理
				redisClient.XAck(ctx, streamName, consumerGroup, message.ID)
			}
		}
	}
}

// Gin 路由（保留 HTTP 接口以兼容旧代码）
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
	initRedis()

	// 启动 HTTP 服务器（用于健康检查和兼容性）
	r := setupRouter()
	go func() {
		port := "8081"
		fmt.Printf("🚀 Event Logger Service HTTP running on http://localhost:%s\n", port)
		r.Run(":" + port)
	}()

	// 启动 Redis Stream 消费者
	fmt.Println("🔄 开始消费 Redis Stream...")
	go consumeFromStream()

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 正在关闭 Event Logger Service...")
	if redisClient != nil {
		redisClient.Close()
	}
	fmt.Println("✅ Event Logger Service 已关闭")
}
