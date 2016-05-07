package sms

import (
	"fmt"
	"github.com/wuyq101/smsverify/config"
	"github.com/wuyq101/smsverify/util"
	"gopkg.in/redis.v3"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type Sms struct {
	redisClient *redis.Client
}

const (
	REDIS_KEY_PREFIX = "sms_verfiy#"
)

func NewSms(redisClient *redis.Client) *Sms {
	return &Sms{
		redisClient: redisClient,
	}
}

//check for frequence limit.
//true --> over limit
func (s *Sms) CheckSendCodeLimit(phone, templateCode string) (bool, error) {
	key := fmt.Sprintf("%ssend_limit_%s_%s", REDIS_KEY_PREFIX, phone, templateCode)
	return s.checkLimit(key, phone, templateCode, config.Instance().SmsFreqLimit)
}

func (s *Sms) CheckVerifyCodeLimit(phone, templateCode string) (bool, error) {
	key := fmt.Sprintf("%sverify_limit_%s_%s", REDIS_KEY_PREFIX, phone, templateCode)
	return s.checkLimit(key, phone, templateCode, config.Instance().SmsFreqLimit)
}

func (s *Sms) checkLimit(key, phone, templateCode string, maxLimit int64) (bool, error) {
	v, err := s.redisClient.Incr(key).Result()
	if err != nil {
		return true, err
	}
	if v == 1 {
		err = s.redisClient.Expire(key, time.Hour).Err()
		if err != nil {
			return true, err
		}
	}
	if v > config.Instance().SmsFreqLimit {
		ttl, _ := s.redisClient.TTL(key).Result()
		if ttl < 0 {
			s.redisClient.Expire(key, time.Hour)
		}
		return true, nil
	}
	return false, nil
}

//generate sms verify code & token
func (s *Sms) GenerateCodeAndToken(phone, templateCode string) (string, string, error) {
	key := fmt.Sprintf("%s_verify_pair_%s_%s", phone, templateCode)
	kv, err := s.redisClient.HGetAllMap(key).Result()
	if err != nil && err != redis.Nil {
		return "", "", err
	}
	var code, token string
	if err == nil && kv != nil {
		token = kv["token"]
	}
	if len(token) == 0 {
		token = util.RandString(16)
	}
	code = s.generateCode()
	if err := s.redisClient.HMSet(key, "code", code, "token", token).Err(); err != nil {
		return "", "", err
	}
	if err := s.redisClient.Expire(key, time.Hour).Err(); err != nil {
		return "", "", err
	}
	return code, token, nil
}

//verify sms code, return status, valid, err
func (s *Sms) VerifyCode(phone, templateCode, token, code string) (string, bool, error) {
	key := fmt.Sprintf("%s_verify_pair_%s_%s", phone, templateCode)
	exists, err := s.redisClient.Exists(key).Result()
	if !exists || err != nil {
		return "code_expire", false, err
	}
	kv, err := s.redisClient.HGetAllMap(key).Result()
	if err != nil {
		return "system_err", false, err
	}
	if token != kv["token"] {
		return "token_invalid", false, nil
	}
	if code != kv["code"] {
		return "code_invalid", false, nil
	}
	//verify pass, del this pair
	go func() {
		sendLimitkey := fmt.Sprintf("%ssend_limit_%s_%s", REDIS_KEY_PREFIX, phone, templateCode)
		verifyLimitkey := fmt.Sprintf("%sverify_limit_%s_%s", REDIS_KEY_PREFIX, phone, templateCode)
		s.redisClient.Del(key, sendLimitkey, verifyLimitkey)
	}()
	return "ok", true, nil
}

func (s *Sms) generateCode() string {
	codeLen := config.Instance().SmsCodeLen
	max := int64(math.Pow(10, float64(codeLen)))
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%0"+strconv.FormatInt(codeLen, 10)+"d", rand.Int63n(max))
}
