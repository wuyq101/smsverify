package sms

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/wuyq101/smsverify/config"
)

type SmsSender interface {
	Send(phone, templateCode, code string, smsParam map[string]string) (bool, error)
}

type AliSmsSender struct {
	AppKey          string
	AppSecret       string
	Format          string
	Version         string
	Method          string
	SignMethod      string
	SmsFreeSignName string
}

type AliSmsOutPut struct {
	Response    *AliSmsRespons    `json:"alibaba_aliqin_fc_sms_num_send_response"`
	ErrResponse *AliSmsErrRespons `json:"error_response"`
}

type AliSmsRespons struct {
	Result AliSmsResult `json:"result"`
}

type AliSmsResult struct {
	ErrCode string `json:"err_code"`
	Model   string `json:"model"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

type AliSmsErrRespons struct {
	Code    int64  `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

func NewAliSmsSender() *AliSmsSender {
	return &AliSmsSender{
		AppKey:          config.Instance().AliAppKey,
		AppSecret:       config.Instance().AliAppSecret,
		Format:          "json",
		Version:         "2.0",
		Method:          "alibaba.aliqin.fc.sms.num.send",
		SignMethod:      "md5",
		SmsFreeSignName: config.Instance().AliSmsFreeSignName,
	}
}

func dialTimeout(network, addr string) (net.Conn, error) {
	timeout := time.Duration(55) * time.Second
	deadline := time.Now().Add(timeout)
	c, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	c.SetDeadline(deadline)
	return c, nil
}

func HttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial:                  dialTimeout,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: time.Duration(55) * time.Second,
		},
	}
}

func (s *AliSmsSender) makeCommonParams() url.Values {
	data := make(url.Values)
	//common params
	data.Add("method", s.Method)
	data.Add("app_key", s.AppKey)
	data.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	data.Add("format", s.Format)
	data.Add("v", s.Version)
	data.Add("sign_method", s.SignMethod)
	//sms common params
	data.Add("sms_type", "normal")
	data.Add("sms_free_sign_name", s.SmsFreeSignName)
	return data
}

func (s *AliSmsSender) Send(phone, templateCode, code string, smsParam map[string]string) (bool, error) {
	//make url params
	data := s.makeCommonParams()
	data.Add("rec_num", phone)
	data.Add("sms_template_code", templateCode)
	//add code to smsParam
	smsParam["code"] = code
	smsCnt, _ := json.Marshal(smsParam)
	data.Add("sms_param", string(smsCnt))
	//sign
	sign := s.Sign(data)
	data.Add("sign", sign)
	//send
	client := HttpClient()
	request, err := http.NewRequest("POST", "http://gw.api.taobao.com/router/rest", bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return false, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("Failed to send sms")
		return false, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	//parse output
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read ali sms response")
		return false, err
	}
	var output AliSmsOutPut
	err = json.Unmarshal(content, &output)
	if err != nil {
		log.WithError(err).Error("Failed to read ali sms response")
		return false, err
	}
	if &output != nil && output.Response != nil {
		log.WithFields(log.Fields{
			"phone":         phone,
			"template_code": templateCode,
			"code":          code,
			"sms_param":     smsParam,
			"response":      string(content),
		}).Info("Send sms success")
		return true, nil
	}
	if &output != nil && output.ErrResponse != nil {
		log.WithFields(log.Fields{
			"phone":         phone,
			"template_code": templateCode,
			"code":          code,
			"sms_param":     smsParam,
			"response":      string(content),
		}).Error("Failed to send sms")
		return false, nil
	}
	return false, nil
}

func (s *AliSmsSender) Sign(data url.Values) string {
	keys := make([]string, 0)
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))
	str := ""
	for _, key := range keys {
		str += key + data.Get(key)
	}
	str = s.AppSecret + str + s.AppSecret
	md5 := md5.New()
	md5.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(md5.Sum(nil)))
}
