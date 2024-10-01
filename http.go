// implement a small http server in this file
package qecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

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
