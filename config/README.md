# 简介

config 是一个基于 go 的 config manager。

# 设计思路

来源于 github.com/astaxie/beego/config

config模块提供了config的相关支持。对于个人而言，学习如何实现一个config handler。  

# 使用方法

## 安装

```
 ?go get github.com/Tobecoder/readygo/config
```
## providers列表

目前包含了`ini`。

## 使用

```
  import (
    "github.com/Tobecoder/readygo/config"
  )
```

使用ini provider

```
  config, err = config.NewConfig("ini", configFile)
  if err != nil {
	//todo
  }
  //todo
```
