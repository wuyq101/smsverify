package config

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	RedisMasterName    string   `json:"redis_master_name"`
	RedisSentinelAddrs []string `json:"redis_sentinel_addrs"`
	SmsFreqLimit       int64    `json:"sms_freq_limit"` //一个手机号每小时最多可以下发的条数
	SmsCodeLen         int64    `json:"sms_code_len"`   //短信验证码长度
}

var conf *Config

func Init(path string) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Failed to load config file")
	}
	conf = &Config{}
	err = json.Unmarshal(buf, conf)
	if err != nil {
		log.WithFields(log.Fields{
			"conf": string(buf),
			"path": path,
			"err":  err,
		}).Fatal("Failed to decode config file")
	}
	if conf.SmsCodeLen <= 0 {
		conf.SmsCodeLen = 6
	}
}

func Instance() *Config {
	if conf == nil {
		Init("./sv.conf")
	}
	return conf
}
