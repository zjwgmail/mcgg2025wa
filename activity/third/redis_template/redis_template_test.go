package redis_template

import (
	"context"
	"go-fission-activity/config"
	"log"
	"testing"
)

func setupRedisConfig() {
	//172.16.100.159:6379
	config.ApplicationConfig.Redis.Address = "182.92.219.71:6379"
	config.ApplicationConfig.Redis.MaxIdle = 16
	config.ApplicationConfig.Redis.MaxActive = 5
	config.ApplicationConfig.Redis.IdleTimeout = 3000
	config.ApplicationConfig.Redis.Password = ""
	config.ApplicationConfig.Redis.Database = 10
}

func TestRedisTemplate_Set(t *testing.T) {
	setupRedisConfig()
	template := NewRedisTemplate()
	log.Println(&template)
	template = NewRedisTemplate()
	log.Println(&template)
	template = NewRedisTemplate()
	log.Println(&template)
	bl := template.Set(context.TODO(), "go_test1111", "abcd")
	t.Logf("set: %v", bl)
}
