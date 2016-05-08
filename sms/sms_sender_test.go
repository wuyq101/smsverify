package sms

import (
	"fmt"
	"net/url"
	"testing"
)

func TestCall(t *testing.T) {
	sender := &AliSmsSender{
		AppSecret: "helloworld",
	}
	data := make(url.Values)
	data.Add("app_key", "12345678")
	data.Add("fields", "num_iid,title,nick,price,num")
	data.Add("format", "json")
	data.Add("method", "taobao.item.seller.get")
	data.Add("num_iid", "11223344")
	data.Add("session", "test")
	data.Add("sign_method", "md5")
	data.Add("timestamp", "2016-01-01 12:00:00")
	data.Add("v", "2.0")
	s := sender.Sign(data)
	if s != "66987CB115214E59E6EC978214934FB8" {
		t.Fatal("sign failed")
	}
	fmt.Println(s)
}
