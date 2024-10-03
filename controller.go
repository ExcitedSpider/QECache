// The controller which handles requests from outside the package
package qecache

import (
	"fmt"
	"log"
	"sync"
)

// the data fetcher. Invoked when cache miss
type Fetcher interface {
	Fetch(key string) ([]byte, error)
}

// Define a function type that implements Getter
// This is useful to allow users pass a anonymous function
type FetcherFunc func(key string) ([]byte, error)

// Define a methods on a function
// This is an interesting golang trick
// This means function is also a normal object
func (f FetcherFunc) Fetch(key string) ([]byte, error) {
	return f(key)
}

type Controller struct {
	// The name of the controller
	// Allow create multiple controllers.
	// The name is used to distinguish them
	name string
	// The data fetcher which is invoked when miss
	fetcher Fetcher
	// the underlying cache structure
	mainCache cache
	// allow loading from peers
	peers PeerDict
}

// global variables
var (
	mu sync.RWMutex
	// keep references to created controllers
	controllers = make(map[string]*Controller)
)

func GetController(name string) *Controller {
	mu.RLock()
	defer mu.RUnlock()
	return controllers[name]
}

func NewController(name string, maxBytes int64, getter Fetcher) *Controller {
	if getter == nil {
		panic("nil getter")
	}

	// controllers is simply a map. We have to protect it in concurrency
	mu.Lock()
	defer mu.Unlock()

	controller := &Controller{name: name, fetcher: getter, mainCache: cache{maxBytes: maxBytes}}
	controllers[name] = controller

	return controller
}

// Get value for a key from cache
func (c *Controller) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := c.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return c.fetch(key)
}

func (c *Controller) RegisterPeers(peers PeerDict) {
	if c.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	c.peers = peers
}

func (c *Controller) fetch(key string) (ByteView, error) {
	if c.peers != nil {
		// there are registered peers
		// we try to fetch from peers first
		// if the key is assigned to current node, ok would be `false`
		// this can correctly force the current node to fetch if missed
		if peer, ok := c.peers.PeerOfKey(key); ok {
			if value, err := c.fetchFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("Cannot hear from peers")
		}
	}
	return c.fetchLocally(key)
}

func (c *Controller) fetchFromPeer(peer RemotePeer, key string) (ByteView, error) {
	bytes, err := peer.Get(c.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{value: bytes}, nil
}

func (c *Controller) fetchLocally(key string) (ByteView, error) {
	bytes, err := c.fetcher.Fetch(key)
	if err != nil {
		return ByteView{}, err

	}

	// we decided to clone the bytes
	// it may not be the most efficient way
	// need to consider in what conditions it is safe to use bytes.
	// A good idea might be that we can have some contract with the user
	clone := make([]byte, len(bytes))
	copy(clone, bytes)

	value := ByteView{value: clone}
	c.populateCache(key, value)
	return value, nil
}

// add some data to the cache manually
// primarily for testing
// it is not recommend to use this in production
func (g *Controller) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
