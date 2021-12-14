# DNSLog-Platform-Golang

相信DNSLog平台已经是安全从业者的标配。而公开的DNSLOG平台域名早已进入流量监控设备的规则库。同时也有隐私问题值得关注。于是撸了（凑了）一个一键搭建Dnslog平台的golang版本。可以使用其一键搭建自己的Dnslog平台。

## 部署

1. 克隆本仓库到你的服务器上

2. 修改配置文件(config.toml)

   ```toml 
   [front]
   template = "index.html"
   [back]
   listenhost = "0.0.0.0"
   listenport = 8000
   domains = [ "ns.bypass.com" ]
   cname = "www.baidu.com"
   [basicauth]
   check = false
   username = "yumu"
   password = "yumusb"
   ```
   如小学英语老师教我们的那样。可以配置
   1. 前端模板文件
   2. 后端监听的主机、端口、域名、与CNAME响应
   3. HTTP BASIC AUTH的是否打开（check=true）与密码配置
   4. 本计划写多域名的，所以domains写成了一个列表

3. 域名准备

   做好NS指向即可。不会检查请看后文

4. 尝试运行


   先执行

   ```shell
   $ go env -w GO111MODULE=on
   $ go env -w GOPROXY=https://goproxy.cn,direct #可选，国内机器不能上github则需要执行此处以设置{代}{理}
   ```

   而后`go run main.go`即可看到如下字样，说明已经可以正常运行。

   ```shell
   [root@centos dnslog] go run main.go 
   2021/12/14 12:30:53 Will cname to  www.baidu.com.
   2021/12/14 12:30:53 OK, Your Dnslog Domain is : ns.bypass.com.
   2021/12/14 12:30:53 Let's Begin!
   2021/12/14 12:30:53 OK, Will listen in  0.0.0.0:8000
   ```
   `go run main.go` 可先在shell前台运行，看功能是否正常使用与检查数据存放目录是否成功创建。如果没问题的话 直接 `nohup go run main.go &`

## 使用

直接web访问进行使用。或者使用api接口。只有两个URI，

1. /new_gen

   生成子域名,返回格式如下：

   ```json
   {
   	"domain":"09fbd867.www.com.",
   	"key":"09fbd867",
   	"token":"iepdbo4yz1vn"
   }
   ```

   token字段为新引进。随机生成的12位字符。而后通过md5运算后取得key作为子域名部分。

   ```go
   token := randSeq(12)
   key := md5sum(token)[0:8]
   ```

   当然你也可以本地进行生成。不过要注意的是所有访问均进行了强制转换为小写，所以你自己本地生成的token要是一个 12 位的小写字符串。

2. /$yourtoken

   通过访问/$yourtoken（此处也就是/iepdbo4yz1vn）可以获取到相关的DNS解析记录。返回为null或者正常数据的JSON形式。


PS：当然，你也可以通过 go build 打包成可执行文件进行跨平台或者离线运行。这都依赖于go的特性。

## 可定义的配置项

详见 config.toml

## 检查NS指向是否成功？

可以通过linux下命令行工具`host -t ns sub.youdomain.com`来确认NS服务器。也可以通过在线工具：https://myssl.com/dns_check.html （选择NS类型）

## Demo

```python
#coding:utf-8
import requests
import json

base = "http://localhost:8000/"
try:
	print("[-] try to get a subdomain.")
	subdomaindata = requests.get(base+"new_gen",timeout=5).json()
	token = subdomaindata['token']
	subdomain = subdomaindata['domain']
	print("[+] this is your subdomain [ %s ], try to resolve it!" % subdomain)
	print("[+] this is your token [ %s ]" % token)
	try:
		requests.get("http://"+subdomain,timeout=2)
	except:
		pass
	data = requests.get(base+token,timeout=5).text
	if(data=="null"):
		print("no data")
	else:
		res = json.loads(data)
		for x in res:
			print(res[x])
except:
	print("error")
```

## 更新日志：


+ 2021/12/14 在log4j2漏洞影响下，破100star。引入http basic auth，改为toml文件修改配置。
+ 2021/4/3 引入token机制，保证隐私性。

+ 1970-2021/4/2 初版本。