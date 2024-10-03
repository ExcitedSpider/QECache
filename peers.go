// Get entry from peers
package qecache

// Keep records of peers in this dictionary
type PeerDict interface {
	SelectPeer(key string) (peer RemotePeer, ok bool)
}

type RemotePeer interface {
	Get(group string, key string) ([]byte, error)
}
