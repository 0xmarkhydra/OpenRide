package extension

import (
	"context"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

type Capability string

const (
	CapabilityPassenger     Capability = "passenger"
	CapabilityParcel        Capability = "parcel"
	CapabilityCarpool       Capability = "carpool"
	CapabilityScheduled     Capability = "scheduled"
	CapabilityCustomerAsset Capability = "customer_asset"
	CapabilityMultiStop     Capability = "multi_stop"
)

// Manifest is intentionally data-first so an operator UI can inspect a service
// without linking itself to the service implementation.
type Manifest struct {
	ID           marketplace.ServiceType `json:"id"`
	Version      string                  `json:"version"`
	DisplayName  string                  `json:"display_name"`
	Description  string                  `json:"description,omitempty"`
	Category     string                  `json:"category"`
	Capabilities []Capability            `json:"capabilities,omitempty"`
	// Contracts maps logical contract names (for example request_schema) to
	// versioned URIs/IDs. This keeps dynamic UI/SDK integration language-neutral.
	Contracts map[string]string `json:"contracts,omitempty"`
}

func (m Manifest) Valid() bool {
	return m.ID != "" && m.Version != "" && m.DisplayName != "" && m.Category != ""
}

// ServiceModule is the extension point for a mobility vertical.
// Core owns marketplace invariants; modules own service-specific validation.
type ServiceModule interface {
	Manifest() Manifest
	ValidateRequest(ctx context.Context, request marketplace.Request) error
}
