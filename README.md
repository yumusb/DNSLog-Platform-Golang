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
   domains = [ "dns.1433.eu.org","dns.bypass.eu.org"]
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

直接web访问进行使用。或者使用api接口。

1. /get_domain

   返回所有可选域名的JSON。对应toml中的配置项
   `[ "dns.1433.eu.org","dns.bypass.eu.org"]`
1. /(new_gen|get_sub_domain)?domain=xxxxx. 

   (为了兼容之前的接口，只好这样了。)  

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
   
   可以指定domain参数（/new_gen?domain=dns.1433.eu.org.，支持GET、POST），传入值必须完全匹配/get_domain接口返回列表之一，不指定或者不匹配则使用列表中的第一个。

2. /($yourtoken|get_results)

   通过访问/$yourtoken（此处也就是/iepdbo4yz1vn）或者访问/get_results并传参（支持GET、POST）12位TOKEN（token=iepdbo4yz1vn）可以获取到相关的DNS解析记录。返回为null或者正常数据的JSON形式。

   可以指定domain参数（/$yourtoken?domain=dns.1433.eu.org.，支持GET、POST），传入值必须完全匹配/get_domain接口返回列表之一，不指定或者不匹配则使用列表中的第一个。不同域名之间的key不通用，内容不覆盖。也就是说在获取子域名时指定了域名，在此处也必须指定域名。


PS：当然，你也可以通过 go build 打包成可执行文件进行跨平台或者离线运行。这都依赖于go的特性。

## 可定义的配置项

详见 config.toml

## 检查NS指向是否成功？

可以通过linux下命令行工具`host -t ns sub.youdomain.com`来确认NS服务器。也可以通过在线工具：https://myssl.com/dns_check.html （选择NS类型）


## 更新日志：
+ 2021/12/18 其他问题1：修改为了Form，兼容GET与POST传参。为获取结果接口增加了固定URL。
+ 2021/12/17 又增加了50star。引入了多域名机制。修改了前端。
+ 2021/12/14 在log4j2漏洞影响下，破100star。引入http basic auth，改为toml文件修改配置。
+ 2021/4/3 引入token机制，保证隐私性。

+ 1970-2021/4/2 初版本。


## 其他问题：

1. 在版本>94的Chrome且使用非https协议访问，并且在GET参数中出现domain 可能出现以下问题
CORS：The request client is not a secure context and the resource is in more-private address space `local`.
可参考 https://developer.chrome.com/blog/private-network-access-update/
2. 可能遇到非预期的退出问题。在`log.咕.com`实际运行中遇到过两次。但是线上环境中做了进程守护把日志给覆盖掉了。在测试环境中数据太小又无法复现。欢迎复现的大哥提交相关日志。
