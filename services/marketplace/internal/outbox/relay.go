package outbox

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/store"
)

type Relay struct {
	Store    *store.Store
	JS       nats.JetStreamContext
	Interval time.Duration
	Batch    int
}

func (r Relay) Run(ctx context.Context) {
	if r.Store == nil || r.JS == nil { return }
	interval := r.Interval
	if interval <= 0 { interval = 500 * time.Millisecond }
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if _, err := r.Store.RelayOutboxBatch(ctx, r.Batch, func(m store.OutboxMessage) error {
			msg := &nats.Msg{Subject: m.Subject, Data: m.Payload, Header: nats.Header{}}
			msg.Header.Set(nats.MsgIdHdr, m.ID)
			_, err := r.JS.PublishMsg(msg)
			return err
		}); err != nil && ctx.Err() == nil {
			log.Printf("marketplace outbox relay: %v", err)
		}
		select {
		case <-ctx.Done(): return
		case <-ticker.C:
		}
	}
}
