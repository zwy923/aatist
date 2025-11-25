package service

import "fmt"

const (
	redisTrendingKey = "community:trending"
)

func likeCountKey(postID int64) string {
	return fmt.Sprintf("community:post:%d:likes", postID)
}

func commentCountKey(postID int64) string {
	return fmt.Sprintf("community:post:%d:comments", postID)
}
