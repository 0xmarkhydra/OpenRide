package realtime

import (
	"encoding/json"
	"sync"
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	ActorID   string    `json:"actor_id,omitempty"`
	TripID    string    `json:"trip_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

type Subscription struct {
	ID      uint64
	ActorID string
	Events  <-chan Event
	close   func()
}

func (s Subscription) Close() { if s.close != nil { s.close() } }

type Hub struct {
	mu      sync.RWMutex
	nextID  uint64
	clients map[string]map[uint64]chan Event
}

func NewHub() *Hub { return &Hub{clients: make(map[string]map[uint64]chan Event)} }

func (h *Hub) Subscribe(actorID string) Subscription {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextID++
	id := h.nextID
	ch := make(chan Event, 32)
	if h.clients[actorID] == nil { h.clients[actorID] = make(map[uint64]chan Event) }
	h.clients[actorID][id] = ch
	return Subscription{ID:id,ActorID:actorID,Events:ch,close:func(){h.unsubscribe(actorID,id)}}
}

func (h *Hub) unsubscribe(actorID string,id uint64){
	h.mu.Lock();defer h.mu.Unlock()
	group:=h.clients[actorID];if group==nil{return}
	if ch,ok:=group[id];ok{delete(group,id);close(ch)}
	if len(group)==0{delete(h.clients,actorID)}
}

func (h *Hub) Publish(actorID string,event Event){
	if actorID==""{return}
	if event.Timestamp.IsZero(){event.Timestamp=time.Now().UTC()}
	event.ActorID=actorID
	h.mu.RLock();defer h.mu.RUnlock()
	for _,ch:=range h.clients[actorID]{
		select{case ch<-event:default:}
	}
}

func Marshal(event Event) []byte { payload,_:=json.Marshal(event); return payload }
