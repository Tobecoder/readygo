# ���

config ��һ������ go �� config manager��

# ���˼·

��Դ�� github.com/astaxie/beego/config

configģ���ṩ��config�����֧�֡����ڸ��˶��ԣ�ѧϰ���ʵ��һ��config handler��  

# ʹ�÷���

## ��װ

```
 ?go get github.com/Tobecoder/readygo/config
```
## providers�б�

Ŀǰ������`ini`��

## ʹ��

```
  import (
    "github.com/Tobecoder/readygo/config"
  )
```

ʹ��ini provider

```
  config, err = config.NewConfig("ini", configFile)
  if err != nil {
	//todo
  }
  //todo
```
