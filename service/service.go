package service

import (
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/wuyq101/smsverify/config"
	"github.com/wuyq101/smsverify/sms"
	"github.com/wuyq101/smsverify/util"
	"gopkg.in/redis.v3"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05",
	})
}

type Service struct {
	router      *httprouter.Router
	redisClient *redis.Client
	sms         *sms.Sms
}

type Handler func(http.ResponseWriter, *http.Request, httprouter.Params) (interface{}, error)

func NewService() *Service {
	config := config.Instance()
	ret := &Service{
		redisClient: redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    config.RedisMasterName,
			SentinelAddrs: config.RedisSentinelAddrs,
			PoolSize:      10,
			DialTimeout:   time.Second,
			ReadTimeout:   time.Second,
			WriteTimeout:  time.Second,
			IdleTimeout:   time.Second * 10,
		}),
		router: httprouter.New(),
	}
	ret.router.PanicHandler = ret.PanicHandler()
	ret.sms = sms.NewSms(ret.redisClient)
	if status := ret.redisClient.Ping(); status.Err() != redis.Nil && status.Err() != nil {
		log.WithFields(log.Fields{
			"redis_master_name":    config.RedisMasterName,
			"redis_sentinel_addrs": config.RedisSentinelAddrs,
		}).Fatal("Failed to start service, can not connect to redis")
	}
	return ret
}

func (s *Service) PanicHandler() func(http.ResponseWriter, *http.Request, interface{}) {
	return func(resp http.ResponseWriter, req *http.Request, rcv interface{}) {
		out := map[string]interface{}{
			"status": "system_err",
			"msg":    "system error",
		}
		outContent, _ := json.Marshal(out)
		log.WithFields(log.Fields{
			"url":            req.RequestURI,
			"method":         req.Method,
			"remote_address": req.RemoteAddr,
			"respone":        string(outContent),
			"err":            rcv,
		}).Error("Failed to process the request")
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write(outContent)

	}
}

func (s *Service) MakeHandler(fn Handler) httprouter.Handle {
	log.WithFields(log.Fields{
		"func_name": s.getFuncName(fn),
	}).Info("Register Handler")
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		startTime := time.Now()
		out, err := fn(resp, req, params)
		if err != nil {
			log.WithError(err).Error("Internal Error")
			panic(err)
		}
		outContent, err := json.Marshal(out)
		if err != nil {
			log.WithError(err).Error("Failed to Marshal output json")
			panic(err)
		}
		resp.Write(outContent)
		log.WithFields(log.Fields{
			"url":            req.RequestURI,
			"method":         req.Method,
			"remote_address": req.RemoteAddr,
			"cost_time":      time.Now().Sub(startTime),
			"respone":        string(outContent),
		}).Info("Finished to process the request")
	}
}

func (s *Service) getFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func (s *Service) Run(port string) {
	s.RegisterHandler()
	log.WithFields(log.Fields{
		"port": port,
	}).Info("Start server for sms verify ...")
	log.Fatal(http.ListenAndServe(":"+port, s.router))
}

func (s *Service) RegisterHandler() {
	s.router.GET("/ping", s.MakeHandler(s.Ping))
	s.router.POST("/sms/code/send", s.MakeHandler(s.SendCode))
	s.router.POST("/sms/code/verify", s.MakeHandler(s.VerifyCode))
}

func (s *Service) Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	return "pong", nil
}

func (s *Service) SendCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	phone := r.FormValue("phone")
	if len(phone) == 0 {
		return s.fail("miss required parameter phone"), nil
	}
	//verify by phone rule
	if !util.ValidatePhone(phone) {
		return s.fail("invalid phone"), nil
	}
	templateCode := r.FormValue("template_code")
	if len(templateCode) == 0 {
		return s.fail("miss required parameter template_code"), nil
	}
	//check send frequence
	overLimit, err := s.sms.CheckSendCodeLimit(phone, templateCode)
	if err != nil {
		log.WithFields(log.Fields{
			"phone": phone,
			"err":   err,
		}).Error("Failed to check send code limit")
		return nil, err
	}
	if overLimit {
		return s.failWithStatus("limit_control", "获取短信验证码过于频繁"), nil
	}
	//generate code and token
	code, token, err := s.sms.GenerateCodeAndToken(phone, templateCode)
	if err != nil {
		log.WithFields(log.Fields{
			"phone": phone,
			"code":  code,
			"token": token,
			"err":   err,
		}).Error("Failed to generate sms verify code")
		return nil, err
	}
	log.WithFields(log.Fields{
		"phone": phone,
		"code":  code,
		"token": token,
	}).Info("Generate sms verify code")
	//TODO send code to phone by third party service
	//return the result
	data := map[string]interface{}{
		"token": token,
	}
	return s.ok("", data), nil
}

func (s *Service) fail(msg string) interface{} {
	return s.failWithStatus("fail", msg)
}

func (s *Service) failWithStatus(status, msg string) interface{} {
	return map[string]interface{}{
		"status": status,
		"msg":    msg,
	}
}

func (s *Service) ok(msg string, data interface{}) interface{} {
	return map[string]interface{}{
		"status": "ok",
		"msg":    msg,
		"data":   data,
	}
}

func (s *Service) VerifyCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) (interface{}, error) {
	phone := r.FormValue("phone")
	if len(phone) == 0 {
		return s.fail("miss required parameter phone"), nil
	}
	//verify by phone rule
	if !util.ValidatePhone(phone) {
		return s.fail("invalid phone"), nil
	}
	templateCode := r.FormValue("template_code")
	if len(templateCode) == 0 {
		return s.fail("miss required parameter template_code"), nil
	}
	token := r.FormValue("token")
	if len(token) == 0 {
		return s.fail("miss required parameter token"), nil
	}
	code := r.FormValue("code")
	if len(code) == 0 {
		return s.fail("miss required parameter code"), nil
	}
	log.WithFields(log.Fields{
		"phone":         phone,
		"template_code": templateCode,
		"token":         token,
		"code":          code,
	}).Info("Start to verify")
	//check verify frequence
	overLimit, err := s.sms.CheckVerifyCodeLimit(phone, templateCode)
	if err != nil {
		log.WithFields(log.Fields{
			"phone":         phone,
			"template_code": templateCode,
			"token":         token,
			"code":          code,
			"err":           err,
		}).Error("Failed to check verify code limit")
		return nil, err
	}
	if overLimit {
		return s.failWithStatus("limit_control", "尝试次数过多"), nil
	}
	status, _, err := s.sms.VerifyCode(phone, templateCode, token, code)
	if err != nil {
		log.WithFields(log.Fields{
			"phone":         phone,
			"template_code": templateCode,
			"token":         token,
			"code":          code,
		}).Error("Failed to verify code")
		return nil, err
	}
	return s.failWithStatus(status, ""), nil
}
