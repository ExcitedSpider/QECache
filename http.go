// Define and implement the required HTTP server and client.
// They delegate all the underlying network operations required by the cache
package qecache

import (
	"QECache/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// ======================================
// HTTP Client
// It mainly used to get entry from peers
// Thus we require it to implement RemotePeer
// ======================================

type httpClient struct {
	baseURL string
}

func (c *httpClient) Get(cname string, key string) ([]byte, error) {
	requestURL := fmt.Sprintf("%v%v/%v",
		c.baseURL,
		url.QueryEscape(cname),
		url.QueryEscape(key),
	)

	res, error := http.Get(requestURL)

	if error != nil {
		return nil, error
	}

	defer res.Body.Close() // remember to always close request body stream

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %v", res.Status)
	}

	bytes, error := io.ReadAll(res.Body)

	if error != nil {
		return nil, fmt.Errorf("error when reading stream: %v", error)
	}

	return bytes, nil
}

// assert httpClient implements RemotePeer (force type check)
// How this trick work?
// for the right hand side, we created a value nil with type (*httpClient)
// then, we assign this value to a variable _ of type RemotePeer.
// we only needs the compiler to carry out type checker, without using
// the variable _
var _ RemotePeer = (*httpClient)(nil)

// ======================================
// HTTP Server
// Serve a set of API that allows peers (and maybe admins) to query what
// the cache holds
// ======================================

// Fixed value shall be declared as constants,
// this can improve performance and prevent unintentional modification
// I prefer to name constants with all capital letters
const DEFAULT_BASE_PATH = "/_cacheserver/"

type HTTPServer struct {
	// selfIP ip address
	// e.g. https://excitedspider.github.io
	selfIP string
	// the path for this server.
	basePath string
	mu       sync.Mutex
	// peer information, used to get entry from peer if cache missed
	peers consistenthash.KeyHashInfo
	// each url has a client.
	// which might not be so efficient but we do it anyway because it's safe
	httpClients map[string]*httpClient
}

type HTTPServerConfig struct {
	// must provide the address
	SelfIP string
	// optional
	BasePath string
}

func NewHTTPServer(config HTTPServerConfig) *HTTPServer {
	if config.SelfIP == "" {
		panic("Must provide Self IP") // TODO: consider automatically get it
	}
	if config.BasePath == "" {
		config.BasePath = DEFAULT_BASE_PATH
	}

	return &HTTPServer{
		selfIP:   config.SelfIP,
		basePath: config.BasePath,
	}
}

func (p *HTTPServer) Log(format string, v ...interface{}) {
	// interface{} means any type

	// the spread operator ... is suffix, weird syntax
	log.Printf("[Server %s %s]: %s", p.basePath, p.selfIP, fmt.Sprintf(format, v...))
}

func (p *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This guard seems useless because I don't know how to handle this situation
	if r == nil {
		panic("Empty Request")
	}

	// TODO: consider allow more than one servers
	// try dispatch to correct ones
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.NotFound(w, r)
		p.Log("User tries to access unknown service: %s", r.URL.Path)
		return
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	// as there is only one defined API, we can do this without parsing the API
	// need to extend it in the future
	p.handleQueryCache(w, r)
}

// Query an cache entry by key
// GET /<basepath>/<controller>/<key>
func (p *HTTPServer) handleQueryCache(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cName := parts[0]
	key := parts[1]

	controller := GetController(cName)
	if controller == nil {
		http.Error(w, "No such controller "+cName, http.StatusNotFound)
		return
	}

	view, err := controller.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: support more content type than plain byte string
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

const DEFAULT_VNODE_SCALAR = 4.

// Set the peers for a server.
// caveat: it removes old peer settings
// Parameters:
// - peerUrls: pass arbitrary peer's urls
func (s *HTTPServer) SetPeers(peerUrls ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peers = *consistenthash.New(DEFAULT_VNODE_SCALAR, nil)
	s.peers.Add(peerUrls...)

	// provide the length so that it might be more efficient
	s.httpClients = make(map[string]*httpClient, len(peerUrls))
	for _, peerUrl := range peerUrls {
		s.httpClients[peerUrl] = &httpClient{baseURL: peerUrl + s.basePath}
	}
}

func (p *HTTPServer) SelectPeer(key string) (RemotePeer, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.selfIP {
		p.Log("Pick peer %s", peer)
		return p.httpClients[peer], true
	}
	return nil, false
}

var _ PeerDict = (*HTTPServer)(nil)
