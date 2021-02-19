# DNSLog-Platform-Golang

相信DNSLog平台已经是安全从业者的标配。而公开的DNSLOG平台域名早已进入流量监控设备的规则库。同时也有隐私问题值得关注。于是撸了（凑了）一个一键搭建Dnslog平台的golang版本。可以使用其一键搭建自己的Dnslog平台。由于是作者的第一个golang程序，难免会有一些小问题。不过 他真的是一键！

## 部署

1. 选择你喜欢的方式下载本目录下的main.go文件

   不多说。

2. 决定是否开放公网访问？

   倒数第几行中的`http.ListenAndServe("localhost:8000", nil) `，这样写的话只能通过localhost进行访问，墙裂建议不要修改，而后通过中间件反向代理后对外开放，方便做访问控制、日志管理等。如果想直接对外开放的话可以修改为:`http.ListenAndServe(":8000", nil)`

3. 域名准备

   做好NS指向即可。不会检查请看后文

4. 尝试运行

   先`go get github.com/miekg/dns` 获取需要的库。`go run main.go yourdomain` 可先在shell前台运行，看功能是否正常使用与检查数据存放目录是否成功创建。如果没问题的话 直接 `nohup go run main.go yourdomain &`

   （第一次运行需要到github拖miekg/dns库，所以需要你的服务器能上github）

## 使用

直接web访问进行使用。或者使用api接口。只有两个URI，

1. /new_gen

   生成随机的八位字符，组合到domain中并返回。当然也可以本地生成，没做限制，无所谓。

2. /八位字符

   第一步生成的八位字符可以直接访问，可以获取到相关的DNS解析记录。返回为null或者正常数据的JSON形式。


PS：当然，你也可以通过 go build 打包成可执行文件进行跨平台或者离线运行。这都依赖于go的特性。

## 可定义的配置项

1. 临时数据存放目录
    ```go
    localdir, _ := os.Getwd()
    tmplogdir = localdir + "/dnslog/" //DNS日志存放目录,可自行更改。
    ```

2. html页面

   HelloHandler方法中的res变量

3. 子域名长度

    数据接口判断了访问路径是否为9,是才会尝试读取文件中储存的记录。

    ```go
    if len(r.URL.Path) == 9 {
    		w.Header().Set("Content-Type", "application/json")
    		res = GetDnslog(r.URL.Path)
    ```

    new_gen接口返回的随机key长度为8

    ```go
    key := randSeq(8)
    ```

    写入文件时判断了子域名长度是否为8

    ```go
    if len(idkey) == 8 {
    ```

## 检查NS指向是否成功？

可以通过linux下命令行工具`host -t ns sub.youdomain.com`来确认NS服务器。也可以通过在线工具：https://myssl.com/dns_check.html （选择NS类型）

## 其他

感谢高学长给予的大力帮助：https://github.com/netchiso

