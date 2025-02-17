package sfu

import (
	"container/list"
	"sync"
)

// buffer ring for cached packets
type packetCaches struct {
	size   int
	mu     sync.RWMutex
	caches *list.List
}

type cachedPacket struct {
	sequence    uint16
	timestamp   uint32
	dropCounter uint16
}

func newPacketCaches(size int) *packetCaches {
	return &packetCaches{
		size:   size,
		mu:     sync.RWMutex{},
		caches: list.New(),
	}
}

func (p *packetCaches) Push(sequence uint16, timestamp uint32, dropCounter uint16) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.caches.PushBack(cachedPacket{
		sequence:    sequence,
		timestamp:   timestamp,
		dropCounter: dropCounter,
	})

	if p.caches.Len() > p.size {
		p.caches.Remove(p.caches.Front())
	}
}

func (p *packetCaches) GetPacket(sequence uint16) (cachedPacket, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

Loop:
	for e := p.caches.Back(); e != nil; e = e.Prev() {
		packet := e.Value.(cachedPacket)
		if packet.sequence == sequence {
			return packet, true
		} else if packet.sequence > sequence {
			break Loop
		}
	}

	return cachedPacket{}, false
}

func (p *packetCaches) GetPacketOrBefore(sequence uint16) (cachedPacket, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for e := p.caches.Back(); e != nil; e = e.Prev() {
		packet := e.Value.(cachedPacket)
		if packet.sequence == sequence || packet.sequence > sequence {
			return packet, true
		}
	}

	return cachedPacket{}, false
}
