package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash"
	"hash/crc64"
	"kortho/p2p/sum"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

func init() {
	gob.Register(net.IP{})
	gob.Register(memberlist.Node{})
}

type P2P interface {
	Run()
	Stop()
	Broadcast([]byte)
	Join([]string) error
}

type Config struct {
	Port          int
	Name          string
	BindAddr      string
	AdvertiseAddr string
}

type NotifyFunc (func(interface{}, []byte))

func New(config Config, u interface{}, notify NotifyFunc) (*p2p, error) {
	p := &p2p{u: u, nf: notify}
	cfg := memberlist.DefaultWANConfig()
	cfg.Events = p
	cfg.Delegate = p
	cfg.Name = config.Name
	cfg.BindPort = config.Port
	cfg.BindAddr = config.BindAddr
	cfg.AdvertisePort = config.Port
	cfg.AdvertiseAddr = config.AdvertiseAddr
	ml, err := memberlist.Create(cfg)
	if err != nil {
		return nil, err
	}
	p.ml = ml
	p.mp = make(map[uint64][]byte)
	p.h = crc64.New(crc64.MakeTable(crc64.ECMA))
	return p, nil
}

func (p *p2p) Join(ids []string) error {
	_, err := p.ml.Join(ids)
	return err
}

func (p *p2p) Broadcast(data []byte) {
	p.Lock()
	defer p.Unlock()
	p.mp[sum.Sum(p.h, data)] = data
}

func (p *p2p) Run() {
	for {
		select {
		case <-p.ch:
			p.ch <- struct{}{}
			return
		case <-time.After(time.Second):
			if ns := p.ml.Members(); len(ns) > 0 {
				p.Lock()
				for _, v := range p.mp {
					for _, n := range ns {
						p.broadcast(n, v)
					}
				}
				p.Unlock()
			}
		}
	}
}

func (p *p2p) Stop() {
	p.ch <- struct{}{}
	<-p.ch
	p.ml.Shutdown()
}

func (p *p2p) broadcast(n *memberlist.Node, data []byte) {
	if n.Name == p.ml.LocalNode().Name {
		return
	}
	nd, err := Encode(*p.ml.LocalNode())
	if err != nil {
		return
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(len(nd)))
	buf = append([]byte("push"), buf...)
	buf = append(buf, nd...)
	p.ml.SendReliable(n, append(buf, data...))
}

func (p *p2p) NotifyJoin(node *memberlist.Node) { fmt.Printf("join: %s\n", node.String()) }

func (p *p2p) NotifyLeave(node *memberlist.Node) { fmt.Printf("leave: %s\n", node.String()) }

func (p *p2p) NotifyUpdate(node *memberlist.Node) { fmt.Printf("update: %s\n", node.String()) }

func (p *p2p) NodeMeta(limit int) []byte {
	return []byte{}
}

func (p *p2p) NotifyMsg(data []byte) {
	switch string(data[:4]) {
	case "push":
		var n memberlist.Node

		data = data[4:]
		len := binary.LittleEndian.Uint64(data)
		data = data[8:]
		if err := Decode(data[:len], &n); err == nil {
			p.recvMsg(&n, data[len:])
		}
	case "pull":
		p.Lock()
		defer p.Unlock()
		delete(p.mp, binary.LittleEndian.Uint64(data[4:]))
	}
}

func (p *p2p) recvMsg(n *memberlist.Node, data []byte) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, sum.Sum(p.h, data))
	p.ml.SendReliable(n, append([]byte("pull"), buf...))
	p.nf(p.u, data)
}

func (p *p2p) GetBroadcasts(overhead, limit int) [][]byte { return nil }

func (p *p2p) LocalState(join bool) []byte { return []byte{} }

func (p *p2p) MergeRemoteState(data []byte, join bool) {}

type p2p struct {
	sync.Mutex
	nf NotifyFunc
	h  hash.Hash64
	u  interface{}
	ch chan struct{}
	mp map[uint64][]byte
	ml *memberlist.Memberlist
}

func Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
