package proxypool

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"notbearparser"
	"sync"
	"time"

	"github.com/dsnet/compress/brotli"
)

type Anonymity int

const (
	ELITE_PROXY Anonymity = iota
	ANONYMOUS
	TRANSPARENT
)

var originURL = "https://free-proxy-list.net/"

var queryString = "#proxylisttable tbody td"

type Proxy struct {
	IP        string
	Port      string
	AreaCode  string
	Anonymity Anonymity
	Https     bool
	Status    bool
}

func NewProxy(ip string, port string, areaCode string, anonymity string, https string) (*Proxy, error) {
	proxy := &Proxy{
		IP:       ip,
		Port:     port,
		AreaCode: areaCode,
	}
	switch anonymity {
	case "elite proxy":
		proxy.Anonymity = ELITE_PROXY
	case "anonymous":
		proxy.Anonymity = ANONYMOUS
	case "transparent":
		proxy.Anonymity = TRANSPARENT
	default:
		return nil, fmt.Errorf("%s is not a valid anonymity param", anonymity)
	}
	switch https {
	case "yes":
		proxy.Https = true
	case "no":
		proxy.Https = false
	default:
		return nil, fmt.Errorf("%s is not a valid https param", https)
	}
	return proxy, nil
}

func (p *Proxy) Check() {
	proxyURL := "%s://%s:%s"
	if p.Https {
		proxyURL = fmt.Sprintf(proxyURL, "https", p.IP, p.Port)
	} else {
		proxyURL = fmt.Sprintf(proxyURL, "http", p.IP, p.Port)
	}
	fixedURL, err := url.Parse(proxyURL)
	if err != nil {
		p.Status = false
		return
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(fixedURL)}, Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://www.autohome.com.cn", nil)
	if err != nil {
		p.Status = false
		return
	}
	for key, value := range autohomeHeaders {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		p.Status = false
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode/200 != 1 {
		p.Status = false
		return
	}
	p.Status = true
	fmt.Printf("proxy pool: %s has passed\n", proxyURL)
}

func (p *Proxy) GetTrans() *http.Transport {
	URL := "%s://%s:%s"
	if p.Https {
		URL = fmt.Sprintf(URL, "https", p.IP, p.Port)
	}
	proxyURL, _ := url.Parse(URL)
	return &http.Transport{Proxy: http.ProxyURL(proxyURL)}
}

func (p *Proxy) GetFixedURL() *url.URL {
	str := fmt.Sprintf("https://%s:%s", p.IP, p.Port)
	URL, _ := url.Parse(str)
	return URL
}

type ProxyList []*Proxy

func NewProxyList() (*ProxyList, error) {
	list := make(ProxyList, 0, 300)
	req, err := http.NewRequest("GET", originURL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range proxyHeaders {
		req.Header.Add(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var reader io.Reader
	switch resp.Header.Get("content-encoding") {
	case "br":
		brReader, err := brotli.NewReader(resp.Body, &brotli.ReaderConfig{})
		if err != nil {
			return nil, err
		}
		reader = brReader
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		reader = gzipReader
	default:
		reader = resp.Body
	}
	byteContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	stringContent := string(byteContent[:])
	parser := notbearparser.NewCursor(stringContent)
	err = parser.Parse()
	if err != nil {
		return nil, err
	}
	tds, err := parser.Search(queryString)
	if err != nil {
		return nil, err
	}
	if len(tds)%8 != 0 {
		return nil, errors.New("td element number error")
	}
	for i := 0; i < len(tds)/8; i++ {
		ip := tds[i*8].Content
		port := tds[i*8+1].Content
		areaCode := tds[i*8+2].Content
		anonymity := tds[i*8+4].Content
		https := tds[i*8+6].Content
		if https == "yes" {
			proxy, err := NewProxy(ip, port, areaCode, anonymity, https)
			if err != nil {
				continue
			}
			list = append(list, proxy)
		}
	}
	var wg sync.WaitGroup
	for _, proxy := range list {
		go func(p *Proxy) {
			wg.Add(1)
			p.Check()
			wg.Done()
		}(proxy)
	}
	wg.Wait()
	newList := make(ProxyList, 0, 300)
	proxy := list.pop()
	for proxy != nil {
		if proxy.Status {
			newList = append(newList, proxy)
		}
		proxy = list.pop()
	}
	return &newList, nil
}

func (pl *ProxyList) pop() *Proxy {
	if len(*pl) == 0 {
		return nil
	}
	proxy, proxyList := (*pl)[0], (*pl)[1:]
	*pl = proxyList
	return proxy
}

func (pl *ProxyList) push(p *Proxy) {
	newList := append(*pl, p)
	*pl = newList
}

type ProxyPool struct {
	Proxys      *ProxyList
	Input       chan *Proxy
	Output      chan *Proxy
	RefreshTime *time.Ticker
	Done        chan interface{}
	Ctx         context.Context
}

func NewProxyPool(ctx context.Context) (*ProxyPool, error) {
	proxyList := make(ProxyList, 0, 300)
	proxyPool := &ProxyPool{
		Proxys:      &proxyList,
		Input:       make(chan *Proxy, 300),
		Output:      make(chan *Proxy, 300),
		RefreshTime: time.NewTicker(30 * time.Minute),
		Ctx:         ctx,
		Done:        make(chan interface{}),
	}
	req, err := http.NewRequest("GET", originURL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range proxyHeaders {
		req.Header.Add(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var reader io.Reader
	switch resp.Header.Get("content-encoding") {
	case "br":
		brReader, err := brotli.NewReader(resp.Body, &brotli.ReaderConfig{})
		if err != nil {
			return nil, err
		}
		reader = brReader
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		reader = gzipReader
	default:
		reader = resp.Body
	}
	byteContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	stringContent := string(byteContent[:])
	parser := notbearparser.NewCursor(stringContent)
	err = parser.Parse()
	if err != nil {
		return nil, err
	}
	tds, err := parser.Search(queryString)
	if err != nil {
		return nil, err
	}
	if len(tds)%8 != 0 {
		return nil, errors.New("td element number error")
	}
	for i := 0; i < len(tds)/8; i++ {
		ip := tds[i*8].Content
		port := tds[i*8+1].Content
		areaCode := tds[i*8+2].Content
		anonymity := tds[i*8+4].Content
		https := tds[i*8+6].Content
		if https == "yes" {
			proxy, err := NewProxy(ip, port, areaCode, anonymity, https)
			if err != nil {
				continue
			}
			*proxyPool.Proxys = append(*proxyPool.Proxys, proxy)
		}
	}
	var wg sync.WaitGroup
	for _, proxy := range *proxyPool.Proxys {
		go func(p *Proxy) {
			wg.Add(1)
			p.Check()
			wg.Done()
		}(proxy)
	}
	wg.Wait()
	newList := make(ProxyList, 0, 300)
	proxy := proxyPool.Proxys.pop()
	for proxy != nil {
		if proxy.Status {
			newList = append(newList, proxy)
		}
		proxy = proxyPool.Proxys.pop()
	}
	proxyPool.Proxys = &newList
	return proxyPool, nil
}

func (pl *ProxyPool) Serve() {
OUTER:
	for {
		select {
		case <-pl.Ctx.Done():
			close(pl.Output)
			pl.RefreshTime.Stop()
			close(pl.Done)
			return
		case proxy := <-pl.Input:
			pl.Proxys.push(proxy)
		case <-pl.RefreshTime.C:
			pl.Refresh()
		default:
			proxy := pl.Proxys.pop()
			if proxy == nil {
				continue OUTER
			}
			pl.Output <- proxy
		}
	}
}

func (pl *ProxyPool) Pop() *Proxy {
	return <-pl.Output
}

func (pl *ProxyPool) Push(proxy *Proxy) {
	pl.Input <- proxy
}

func (pl *ProxyPool) Refresh() {
	newList, err := NewProxyList()
	if err != nil {
		return
	}
	pl.Proxys = newList
}
