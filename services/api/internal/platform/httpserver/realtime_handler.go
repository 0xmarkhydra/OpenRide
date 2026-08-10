package httpserver

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"

	"flashx/services/api/internal/realtime"
)

func (s *Server) realtimeSocket(w http.ResponseWriter, r *http.Request) {
	if s.deps.Auth == nil || s.deps.Realtime == nil {
		writeError(w, http.StatusServiceUnavailable, "REALTIME_UNAVAILABLE", "Realtime service is unavailable", nil)
		return
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Bearer authentication is required", nil)
		return
	}
	claims, err := s.deps.Auth.Authenticate(strings.TrimSpace(authorization[len("Bearer "):]))
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionContextTakeover})
	if err != nil { return }
	defer conn.CloseNow()

	sub := s.deps.Realtime.Subscribe(claims.ActorID)
	defer sub.Close()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := conn.Write(ctx, websocket.MessageText, realtime.Marshal(realtime.Event{
		Type: "session.ready", Data: map[string]any{"actor_id": claims.ActorID, "role": claims.Role},
	})); err != nil { return }

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			_, _, err := conn.Read(ctx)
			if err != nil { return }
		}
	}()

	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done(): return
		case <-readDone: return
		case event, ok := <-sub.Events:
			if !ok { return }
			if err := conn.Write(ctx, websocket.MessageText, realtime.Marshal(event)); err != nil { return }
		case <-ping.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pingCtx)
			pingCancel()
			if err != nil { return }
		}
	}
}
