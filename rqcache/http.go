package rqcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"rqcache/consistenthash"
	pb "rqcache/rqcachepb"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
)

// 默认的基本路径和副本数量
const (
	defaultBasePath = "/_rqcache/"
	defaultReplicas = 50
)

// HTTPPool 实现了 PeerPicker 接口，用于管理一组 HTTP 同伴节点。
type HTTPPool struct {
	self        string                 // 当前节点的基本 URL，例如 "https://example.net:8000"
	basePath    string                 // HTTP 路由的基本路径
	mu          sync.Mutex             // 保护 peers 和 httpGetters 的并发访问
	peers       *consistenthash.Map    // 一致性哈希算法的实例，用于选择节点
	httpGetters map[string]*httpGetter // 同伴节点的 HTTP 客户端，按格式 "http://10.0.0.2:8008" 进行索引
}

// NewHTTPPool 初始化一个 HTTP 同伴节点池。
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 使用服务器名称记录信息
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[服务器 %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有 HTTP 请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool 服务意外的路径：" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> 是必需的
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "错误的请求", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "没有该组："+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将值以 proto 消息的形式写入响应体
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set 更新节点池的同伴列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据键选择一个同伴节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("选择同伴节点 %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// PeerPicker 接口的实现，表示可以选择一个同伴节点
var _ PeerPicker = (*HTTPPool)(nil)

// httpGetter 实现了 PeerGetter 接口，表示通过 HTTP 获取值的客户端
type httpGetter struct {
	baseURL string
}

// Get 通过 HTTP 获取值的实现
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应体：%v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("解码响应体：%v", err)
	}

	return nil
}

// PeerGetter 接口的实现，表示可以通过 HTTP 获取值
var _ PeerGetter = (*httpGetter)(nil)
