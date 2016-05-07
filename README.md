# Sms verify

## API

版本|时间|作者|注释
----|---|---|---
v0.1 | 2016-05-04 | YingqiangWu | 短信验证码服务第一版

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
params | Y | json | 发送验证码短信所需的其他参数，json格式，具体参数由短信通道决定

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
limit_control | 触发流控限制，短时间内同一手机多次获取验证码 
token_invalid | token值不正确
code_invalid | 验证码值不正确
code_expire | 验证码已经失效
