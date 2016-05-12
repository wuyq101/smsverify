# Sms verify

版本|时间|作者|注释
----|---|---|---
v0.1 | 2016-05-04 | YingqiangWu | 短信验证码服务第一版

## Deploy

### build
cd $GOPATH/src  
git clone https://github.com/wuyq101/smsverify.git  
cd smsverify  
go build  

### run
 ./smsverify --conf=sv.conf  --port=8081  
 
### Configuration

key |  type | desc
--- | ----- | ----
redis_master_name | string | redis master name: mymaster
redis_sentinel_addrs | string\[\] | redis sentinel addres : ["127.0.0.1:26379"]
sms_freq_limit | int | 一个手机号一个小时之内错误次数上限，多次重新获取验证码或者多次错误验证，都受这个值现在，如果有一次正确验证，之前的错误次数会清零
sms_code_len | int | 短信验证码的长度
ali_app_key | string | 阿里大鱼短信的key
ali_app_secret | string | 阿里大鱼短信的秘钥
ali_sms_free_sign_name | string | 阿里大鱼短信的签名 



## API

### 通用status说明

错误码|错误描述
------|-------
system_err | 系统错误 
ok | 成功  
fail | 失败,具体原因见返回的msg


### 发送验证码 
POST /sms/code/send

参数|必需|类型|说明
----|----|----|----
phone | Y | string | 接收验证码的手机号码
template_code | Y | string | 短信模板号
sms_param | Y | json | 发送验证码短信所需的其他参数，json格式，具体参数由短信通道决定

成功时候返回结果示例：

	{
		"status" : "ok",
		"msg" : "",
		"data" : {
			"token" : "FgaYoYDhaFZjHgzS"
		}
	}

####status解释

错误码|错误描述
------|-------
limit_control | 触发流控限制，短时间内同一手机多次获取验证码 
sms_server_err | 短信通道异常，发送失败




### 验证短信验证码
POST /sms/code/verify

参数|必需|类型|说明
----|----|----|----
phone | Y | string | 待验证的的手机号
template_code | Y | string | 之前发送验证码时候的短信模板号
token | Y | string | 之前发验证码时候返回的token值
code | Y | string | 用户提交的验证码

成功时返回结果示例：

	{
		"status" : "ok",
		"msg" : "ok",
	}
	
####status解释

错误码|错误描述
------|-------
limit_control | 触发流控限制，短时间内同一手机多次尝试验证 
token_invalid | token值不正确
code_invalid | 验证码值不正确
code_expire | 验证码已经失效
