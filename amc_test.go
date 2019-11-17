package acm

import (
	"context"
	"fmt"
	"go-ctf/common/paladin"
	"os"
	"testing"
	"time"
)
var clinet paladin.Client
func TestACM(t *testing.T) {
	os.Setenv("ACM_ZONE_ID", "cn-shanghai")
	os.Setenv("ACM_ENDPOINT_ADDR", "acm.aliyun.com")
	os.Setenv("ACM_NAMESPACE_ID", "")
	os.Setenv("ACCESS_KEY", "")
	os.Setenv("SECRET_KEY", "")
	os.Setenv("ACM_GROUP", "DEFAULT_GROUP")
	os.Setenv("ACM_DATA", "http,app")
	acm := &acmDriver{}
	c, err := acm.New()
	if err != nil {
		panic(err)
		return
	}
	clinet = c
}

func TestGet(t *testing.T) {
	str, _ := clinet.Get("http").String()
	fmt.Printf("HTTP Config: %s", str)
}

func TestGetAll(t *testing.T) {
	str1, _ := clinet.GetAll().Get("http").String()
	fmt.Printf("HTTP Config: %s", str1)
}

func TestWatchEvent(t *testing.T) {
	go func() {
		for event := range clinet.WatchEvent(context.Background(), "app") {
			fmt.Printf("app: %s", event.Value)
		}
	}()
	time.Sleep(20 *time.Second)
	clinet.Close()
}