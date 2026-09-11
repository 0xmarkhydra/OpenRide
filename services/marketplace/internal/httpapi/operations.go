package httpapi

import (
	"context"
	"errors"
	"net/http"

	marketplacestore "github.com/0xmarkhydra/OpenRide/services/marketplace/internal/store"
)

type outboxStatusReader interface {
	OutboxStatus(context.Context) (marketplacestore.OutboxStatus, error)
}

func (s *Server) outboxStatus(w http.ResponseWriter, r *http.Request) {
	reader, ok := s.v2.(outboxStatusReader)
	if !ok {
		writeError(w, http.StatusNotImplemented, "OUTBOX_STATUS_UNAVAILABLE", "outbox status is unavailable for this store")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2_000_000_000)
	defer cancel()
	status, err := reader.OutboxStatus(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "OUTBOX_STATUS_FAILED", "outbox status could not be read")
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: status})
}

func outboxStatusFromBase(ctx context.Context, base V2Store) (marketplacestore.OutboxStatus, error) {
	reader, ok := base.(outboxStatusReader)
	if !ok {
		return marketplacestore.OutboxStatus{}, errors.New("outbox status unsupported by base store")
	}
	return reader.OutboxStatus(ctx)
}
