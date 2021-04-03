package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ip string
var tmplogdir string

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func checkdir() {
	localdir, _ := os.Getwd()
	tmplogdir = localdir + string(os.PathSeparator)+"dnslog"+string(os.PathSeparator) //DNS日志存放目录,可自行更改。
	if !Exists(tmplogdir) {
		log.Print("Path `" + tmplogdir + " `is not exists,will try to create")
		err := os.MkdirAll(tmplogdir, 0666)
		if err != nil {
			fmt.Println(err)
			log.Fatal("Path `" + tmplogdir + " create fail. Please Create It.")
		}
	}
}
func md5sum(str string) string  {
    h := md5.New()
    h.Write([]byte(str))
    return hex.EncodeToString(h.Sum(nil))
}
var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
var topDomain string

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetDnslog(id string) string {
	content := "content"
	path := tmplogdir + string(os.PathSeparator) + id
	if Exists(path) {
		file, _ := os.Open(path)
		defer file.Close()
		tmpcontent, _ := ioutil.ReadAll(file)
		content = string(tmpcontent)
		res := make(map[int]map[string]string)
		data := make(map[string]string)
		i := 0
		y := []string{}
		for _, x := range strings.Split(content, "\n") {
			if len(x) > 5 {
				y = strings.Split(x, "|")
				data["time"] = y[0]
				data["ip"] = y[1]
				data["subdomain"] = y[2]
				res[i] = data
				i++
				data = make(map[string]string)
			}

		}
		enc, _ := json.Marshal(res)
		content = string(enc)
	} else {
		content = "null"
	}

	return string(content)

}
func HelloHandler(w http.ResponseWriter, r *http.Request) {

	res := "Hello World"
	if len(r.URL.Path) == 13 {
		w.Header().Set("Content-Type", "application/json")
		id := md5sum(strings.ToLower(r.URL.Path)[1:13])[0:8]
		res = GetDnslog(id)
		
	} else if r.URL.Path == "/new_gen" {
		w.Header().Set("Content-Type", "application/json")
		rand.Seed(time.Now().UnixNano())
		token := randSeq(12)
		key := md5sum(token)[0:8]
		data := make(map[string]string)
		data["token"] = token
		data["key"] = key
		data["domain"] = key + "." + topDomain
		enc, _ := json.Marshal(data)
		res = string(enc)
		
	} else {
		res = `<!DOCTYPE html><html><head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"><title>DNSLOG Platform</title><meta name="keywords" content="dnslog,dnslog平台"><meta name="description" content="一个无需注册就可以快速使用的DNSLog平台"><style>td{text-align:center;margin:auto}#domainarea p{display:inline-block}</style></head><body><div id="header" style="text-align:center;padding-top:2%%"><p style="font-size:30px">DNSLOG平台</p><hr style="height:2px;border:none;border-top:2px dashed #87cefa"><br></div><script>function getCookie(e){for(var t=e+"=",n=document.cookie.split(";"),o=0;o<n.length;o++){var r=n[o].trim();if(0==r.indexOf(t))return r.substring(t.length,r.length)}return""}function GetDomain(){if(key=getCookie("key"),""!=key&&1!=confirm("获取新的子域名后将会丢失 "+key+"，请注意保存"))return!1;var e=window.XMLHttpRequest?new XMLHttpRequest:new ActiveXObject("Microsoft.XMLHTTP");e.responseType="json",e.onreadystatechange=function(){4==e.readyState&&200==e.status&&(document.cookie="key="+e.response.domain,document.cookie="token="+e.response.token,document.getElementById("myDomain").innerHTML=e.response.domain,document.getElementById("token").innerHTML=e.response.token)},e.open("GET","/new_gen?t="+Math.random(),!0),e.send()}function GetRecords(){var n=window.XMLHttpRequest?new XMLHttpRequest:new ActiveXObject("Microsoft.XMLHTTP");n.onreadystatechange=function(){if(4==n.readyState&&200==n.status){var e=n.responseText;if(""==e||null==e||"null"==e)ktable='<tr bgcolor="#ADD3EF"><th width="45%%">DNS Query Record</th><th width="30%%">IP Address</th><th width="25%%">Created Time</th></tr><td colspan="3" align="center">No Data</td>',document.getElementById("myRecords").innerHTML=ktable;else{obj=JSON.parse(e),table='<tr bgcolor="#ADD3EF"><th width="45%%">DNS Query Record</th><th width="30%%">IP Address</th><th width="25%%">Created Time</th></tr>';for(var t=Object.keys(obj).length-1;t>=(0<Object.keys(obj).length-10?Object.keys(obj).length-10:0);t--)table=table+"<tr><td>"+obj[t].subdomain+"</td><td>"+obj[t].ip+"</td><td>"+obj[t].time+"</td></tr>";document.getElementById("myRecords").innerHTML=table}}},n.open("GET","/"+getCookie("token")+"?t="+Math.random(),!0),n.send()}</script><div id="content" style="text-align:center"><button type="button" onclick="GetDomain()">Get SubDomain</button><button type="button" onclick="GetRecords()">Refresh Record</button><br><div id="domainarea">🌐:<p id="myDomain"></p>&nbsp;&nbsp;🔑:<p id="token"></p></div><center><table id="myRecords" width="700" border="0" cellpadding="5" cellspacing="1" bgcolor="#EFF3FF" style="word-break:break-all;word-wrap:break-all"><tbody><tr bgcolor="#ADD3EF"><th width="45%%">DNS Query Record</th><th width="30%%">IP Address</th><th width="25%%">Created Time</th></tr><tr><td colspan="3" align="center">No Data</td></tr></tbody></table></center></div><script>key=getCookie("key"),token=getCookie("token"),""!=key&&""!=token&&(document.getElementById("myDomain").innerHTML=key,document.getElementById("token").innerHTML=token,GetRecords())</script><div style="text-align:center;margin:0 auto;bottom:100px;width:99.6%%;padding-top:3%%"><hr style="height:2px;border:none;border-top:2px dashed #87cefa"><br><center><span style="color:#add3ef">Copyright&copy;2021 DNSLOG Platform All Rights Reserved.</span></center></div></body></html>`
	}
	fmt.Fprintf(w, res)
}

type Tunnel struct {
	Messages       chan string
	cancel         chan struct{}
	fgListsLock    sync.Mutex
	topDomain      string
	domains        chan string
	maxMessageSize int
}

func NewTunnel(topDomain string, expiration time.Duration, maxMessageSize int) *Tunnel {
	tun := &Tunnel{
		Messages:       make(chan string, 256),
		cancel:         make(chan struct{}),
		topDomain:      topDomain,
		domains:        make(chan string, 256),
		maxMessageSize: maxMessageSize,
	}
	go tun.listenDomains()
	return tun
}

func (tun *Tunnel) Close() {
	close(tun.cancel)
}

func (tun *Tunnel) listenDomains() {
	for {
		select {
		case <-tun.cancel:
			return
		case domain := <-tun.domains:
			func() {
				tun.fgListsLock.Lock()
				defer tun.fgListsLock.Unlock()
				domain = strings.ToLower(domain)
				if strings.Contains(domain,"."+tun.topDomain){
					idkeys := strings.Split(domain[0:len(domain)-len(tun.topDomain)-1], ".")
					idkey := idkeys[len(idkeys)-1]
					//log.Print(idkey)
					if len(idkey) == 8 {
						fd, _ := os.OpenFile(tmplogdir+idkey, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
						fd_time := time.Now().Format("2006-01-02 15:04:05")
						fd_content := strings.Join([]string{fd_time, "|", ip, "|", domain, "\n"}, "")
						log.Print(fd_content)
						buf := []byte(fd_content)
						fd.Write(buf)
						fd.Close()
					}
				}
				
			}()
		}
	}
}

// ServeDNS handles DNS queries, records them, and replies with a CNAME to blackhole-1.iana.org.
func (tun *Tunnel) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) < 1 {
		return
	}
	ip = w.RemoteAddr().String()
	domain := r.Question[0].Name
	if r.Question[0].Qtype == dns.TypeA {
		tun.domains <- domain
	}

	m := &dns.Msg{}
	m.SetReply(r)
	m.Answer = []dns.RR{
		&dns.CNAME{
			Hdr:    dns.RR_Header{Name: domain, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 0},
			Target: "blackhole-1.iana.org.",
		},
	}
	err := w.WriteMsg(m)
	if err != nil {
		log.Println(err)
	}
}

func main() {

	port := flag.Int("port", 53, "port to run on")
	expiration := flag.Int("expiration", 60, "seconds an incomplete message is retained before it is deleted")
	maxMessageSize := flag.Int("maxMessageSize", 5000, "maximum encoded size (in bytes) of a message")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Dnslog Platform requires a domain name parameter, such as `dns1.tk` or `go.dns1.tk`, And check your domain's ns server point to this server")
	}
	checkdir()
	topDomain = dns.Fqdn(strings.ToLower(flag.Arg(0)))
	expirationDuration := time.Duration(*expiration) * time.Second
	tun := NewTunnel(topDomain, expirationDuration, *maxMessageSize)
	dns.Handle(topDomain, tun)
	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()
	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "tcp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set tcp listener %s\n", err.Error())
		}
	}()
	log.Print("Everything is ok, Let's Begin")
	http.HandleFunc("/", HelloHandler)
	http.ListenAndServe("localhost:8000", nil)
	select {} // block foreve
}
