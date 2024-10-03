// Get entry from peers
package qecache

// Select a peer according to the key (by hashing)
type PeerSelector interface {
	SelectPeer(key string) (peer RemotePeer, ok bool)
}

type RemotePeer interface {
	Get(group string, key string) ([]byte, error)
}
