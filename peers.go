// Get entry from peers
package qecache

// Keep records of peers in this dictionary
type PeerDict interface {
	PeerOfKey(key string) (peer RemotePeer, ok bool)
}

type RemotePeer interface {
	Get(namespace string, key string) ([]byte, error)
}
