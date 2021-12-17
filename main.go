package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

var ip string
var tmplogdir string

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func checkdir(path string) {
	logdir := ""
	if path == tmplogdir {
		logdir = tmplogdir + string(os.PathSeparator)
	} else {
		logdir = tmplogdir + string(os.PathSeparator) + path + string(os.PathSeparator)
	}
	if !Exists(logdir) {
		log.Print("Path `" + logdir + " `is not exists,will try to create")
		err := os.MkdirAll(logdir, 0666)
		if err != nil {
			fmt.Println(err)
			log.Fatal("Path `" + logdir + " create fail. Please Create It.")
			os.Exit(1)
		}
	}
}
func md5sum(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
var topDomain []string

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetDnslog(id string, domain string) string {
	content := "content"
	path := tmplogdir + string(os.PathSeparator) + domain + string(os.PathSeparator) + id
	if Exists(path) {
		file, _ := os.Open(path)
		defer file.Close()
		tmpcontent, _ := ioutil.ReadAll(file)
		content = string(tmpcontent)
		res := make(map[int]map[string]string)
		data := make(map[string]string)
		i := 0
		var y []string
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

	if dnslogserver.Basicauth.Check {
		u, p, ok := r.BasicAuth()
		if !ok {
			log.Println("Error parsing basic auth")
			w.Header().Set("WWW-Authenticate", `Basic realm="My REALM"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if u != dnslogserver.Basicauth.Username || p != dnslogserver.Basicauth.Password {
			log.Println("Basic auth Failed", u)
			w.Header().Set("WWW-Authenticate", `Basic realm="My REALM"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	res := "Hello World"
	if len(r.URL.Path) == 13 || r.URL.Path == "/get_results" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-control", "no-store")
		r.ParseForm()
		key := "00000000"
		if len(r.URL.Path) == 13 {
			key = md5sum(strings.ToLower(r.URL.Path)[1:13])[0:8]
		}
		if len(r.Form["token"]) > 0 && len(r.Form["token"][0]) == 12 {
			key = md5sum(strings.ToLower(r.Form["token"][0]))[0:8]
		}
		domain := topDomain[0]
		if len(r.Form["domain"]) > 0 && r.Form["domain"][0] != "" {
			for _, tmpdomain := range topDomain {
				if strings.ToLower(r.Form["domain"][0]) == tmpdomain {
					domain = tmpdomain
					break
				}
			}
		}

		res = GetDnslog(key, domain)
	} else if r.URL.Path == "/new_gen" || r.URL.Path == "/get_sub_domain" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-control", "no-store")
		rand.Seed(time.Now().UnixNano())
		token := randSeq(12)
		key := md5sum(token)[0:8]
		data := make(map[string]string)
		data["domain"] = key + "." + topDomain[0]
		r.ParseForm()
		if len(r.Form["domain"]) > 0 && r.Form["domain"][0] != "" {

			for _, tmpdomain := range topDomain {
				if strings.ToLower(r.Form["domain"][0]) == tmpdomain {
					data["domain"] = key + "." + tmpdomain
					break
				}
			}
		}

		data["token"] = token
		data["key"] = key

		enc, _ := json.Marshal(data)
		res = string(enc)

	} else if r.URL.Path == "/get_domain" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-control", "no-store")
		enc, _ := json.Marshal(topDomain)
		res = string(enc)
	} else {
		res = templatehtml
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
	w.Header().Set("X-Powered-By", "https://github.com/yumusb/DNSLog-Platform-Golang")
	w.Write([]byte(res))

}

type Tunnel struct {
	Messages       chan string
	cancel         chan struct{}
	fgListsLock    sync.Mutex
	topDomain      []string
	domains        chan string
	maxMessageSize int
}

func NewTunnel(topDomain []string, expiration time.Duration, maxMessageSize int) *Tunnel {
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
				//domain = strings.ToLower(domain)
				for _, tmpdomain := range tun.topDomain {
					if strings.Contains(domain, "."+tmpdomain) {
						idkeys := strings.Split(domain[0:len(domain)-len(tmpdomain)-1], ".")
						idkey := idkeys[len(idkeys)-1]
						//log.Print(idkey)
						if len(idkey) == 8 {
							idkey = strings.ToLower(idkey)
							fd, _ := os.OpenFile(tmplogdir+string(os.PathSeparator)+tmpdomain+string(os.PathSeparator)+idkey, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
							fd_time := time.Now().Format("2006-01-02 15:04:05")
							fd_content := strings.Join([]string{fd_time, "|", ip, "|", domain, "\n"}, "")
							log.Print(fd_content)
							buf := []byte(fd_content)
							fd.Write(buf)
							fd.Close()
						}
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
			Target: cname,
		},
	}
	err := w.WriteMsg(m)
	if err != nil {
		log.Println(err)
	}
}

type server struct {
	Backend   back  `toml:"back"`
	Frontend  front `toml:"front"`
	Basicauth basic `toml:"basicauth"`
}
type front struct {
	Template string `toml:"template"`
}
type back struct {
	Listenhost string   `toml:"listenhost"`
	Listenport int      `toml:"listenport"`
	Domains    []string `toml:"domains"`
	Cname      string   `toml:"cname"`
}

type basic struct {
	Check    bool   `toml:"check"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var dnslogserver server
var templatehtml string
var cname string

func main() {

	configfile := "config.toml"
	if _, err := toml.DecodeFile(configfile, &dnslogserver); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	//log.Println(dnslogserver.Backend.Domains)

	cname = dns.Fqdn(dnslogserver.Backend.Cname)
	log.Println("Will cname to ", cname)

	content, err := ioutil.ReadFile(dnslogserver.Frontend.Template)
	if err != nil {
		panic(err)
	}
	templatehtml = string(content)
	port := 53
	expiration := 60
	maxMessageSize := 5000
	for _, tmpdomain := range dnslogserver.Backend.Domains {
		topDomain = append(topDomain, dns.Fqdn(strings.ToLower(tmpdomain)))
	}
	//topDomain = dns.Fqdn(strings.ToLower(dnslogserver.Backend.Domains[0]))
	log.Println("OK, Your Dnslog Domain is :", topDomain)

	if dnslogserver.Basicauth.Check {
		log.Println("BasicAuth is open")
		log.Println("BasicAuth Username:", dnslogserver.Basicauth.Username)
		log.Println("BasicAuth Password:", dnslogserver.Basicauth.Password)
	}
	expirationDuration := time.Duration(expiration) * time.Second
	tun := NewTunnel(topDomain, expirationDuration, maxMessageSize)

	tmplogdir = "dnslog"
	checkdir(tmplogdir)
	for _, tmpdomain := range topDomain {
		dns.Handle(tmpdomain, tun)
		checkdir(tmpdomain)
	}
	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()
	go func() {
		srv := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "tcp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set tcp listener %s\n", err.Error())
		}
	}()
	log.Print("Let's Begin!")
	fs := http.FileServer(http.Dir("static"))
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", HelloHandler)
	listenserver := dnslogserver.Backend.Listenhost + ":" + strconv.Itoa(dnslogserver.Backend.Listenport)
	log.Println("OK, Will listen in ", listenserver)
	http.ListenAndServe(listenserver, mux)
	select {} // block foreve
}
