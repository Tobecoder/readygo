# 简介

session 是一个基于 go 的 session manager。

# 设计思路

来源于 github.com/astaxie/beego/session

session模块提供了session的相关支持。对于个人而言，学习如何实现一个session handler。  

# 使用方法

## 安装

```
  go get github.com/Tobecoder/readygo/session
```
## providers列表

目前包含了memory。

## 使用

```
  import (
    "github.com/Tobecoder/readygo/session"
  )
```

在应用中定义全局的管理器
```
  var globalSessions *session.Manager
```

使用memory provider

```
  func init() {
  	globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessid","gclifetime":3600}`)
  	go globalSessions.GC()
  }
```
