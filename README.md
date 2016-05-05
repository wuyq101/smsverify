# Sms verify

## API

| 版本 | 时间 | 作者 | 注释|
|-- | -- | -- | --|
|v0.1 | 2016-05-04 | YingqiangWu | 短信验证码服务第一版|

### 发送验证码 
GET /sms/code/send

|参数 | 必需 | 类型 | 说明|
|-- | -- | -- | --|
|phone | Y | string | 接收验证码的手机号码|
|params | Y | json | 发送验证码短信所需的其他参数，json格式，具体参数由短信通道决定|

成功时候返回结果示例：

    {
        "status" : "ok",
        "msg" : "ok",
        "data" : {
            "token" : "234sdfsdjksdiulskdfjskd"
        }
    }


### 验证短信验证码
GET /sms/code/verify

|参数|必需|类型|说明|
|--|--|--|--|
|phone | Y | string | 有效的手机号|
|token | Y | string | 之前发验证码时候返回的token值|
|code | Y | string | 待验证的手机号|

成功时返回结果示例：

    {
        "status" : "ok",
        "msg" : "ok",
    }
    

