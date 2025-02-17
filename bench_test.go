package connmgr

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/libp2p/go-libp2p-core/network"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"

	"github.com/libp2p/go-libp2p-core/protocol"

)

func randomConns(tb testing.TB) (c [5000]network.Conn) {
	for i := range c {
		c[i] = randConn(tb, nil)
	}
	return c
}

func BenchmarkLockContention(b *testing.B) {
	conns := randomConns(b)
	ps := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewProtoBook(), pstoremem.NewPeerMetadata())
	cm := NewConnManager(1000, 1000, 0, ps, map[protocol.ID]int{})
	not := cm.Notifee()

	kill := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-kill:
					return
				default:
					cm.TagPeer(conns[rand.Intn(len(conns))].RemotePeer(), "another-tag", 1)
				}
			}
		}()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rc := conns[rand.Intn(len(conns))]
		not.Connected(nil, rc)
		cm.TagPeer(rc.RemotePeer(), "tag", 100)
		cm.UntagPeer(rc.RemotePeer(), "tag")
		not.Disconnected(nil, rc)
	}
	close(kill)
	wg.Wait()
}
