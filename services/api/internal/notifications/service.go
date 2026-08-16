package notifications

import (
	"context"
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
)

type Service struct {
	store    Store
	provider Provider
	now      func() time.Time
}

func NewService(store Store, provider Provider) *Service {
	if provider == nil {
		provider = DisabledProvider{}
	}
	return &Service{store: store, provider: provider, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) ProviderName() string { return s.provider.Name() }

func (s *Service) RegisterDevice(actorID string, role Role, input RegisterDeviceInput) (Device, error) {
	actorID = strings.TrimSpace(actorID)
	token := strings.TrimSpace(input.Token)
	if actorID == "" || !validRole(role) || !validPlatform(input.Platform) || len(token) < 16 || len(token) > 4096 {
		return Device{}, ErrInvalidInput
	}
	now := s.now()
	return s.store.UpsertDevice(Device{
		ID: ids.New("pushdev"), ActorID: actorID, ActorRole: role, Platform: input.Platform,
		Token: token, Enabled: true, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) DisableDevice(actorID string, role Role, deviceID string) error {
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(deviceID) == "" || !validRole(role) {
		return ErrInvalidInput
	}
	return s.store.DisableDevice(actorID, role, deviceID)
}

func (s *Service) Enqueue(actorID string, role Role, eventType, resourceID string) error {
	title, body, ok := renderMessage(role, eventType)
	if !ok || strings.TrimSpace(actorID) == "" {
		return nil
	}
	now := s.now()
	return s.store.Enqueue(Message{
		ID: ids.New("ntf"), ActorID: actorID, ActorRole: role, EventType: eventType,
		ResourceID: resourceID, Title: title, Body: body,
		Data:   map[string]any{"event_type": eventType, "resource_id": resourceID},
		Status: StatusPending, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) ProcessPending(ctx context.Context, limit int) error {
	messages, err := s.store.ListPending(limit)
	if err != nil {
		return err
	}
	for _, message := range messages {
		if err := s.processOne(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processOne(ctx context.Context, message Message) error {
	devices, err := s.store.ListDevices(message.ActorID, message.ActorRole)
	if err != nil {
		return err
	}
	message.Attempts++
	message.UpdatedAt = s.now()
	if len(devices) == 0 {
		message.Status = StatusSkipped
		message.LastError = "no_active_device"
		return s.store.SaveMessage(message)
	}

	sent := 0
	lastError := ""
	for _, device := range devices {
		if err := s.provider.Send(ctx, device, message); err != nil {
			if errors.Is(err, ErrProviderDisabled) {
				lastError = "provider_disabled"
				continue
			}
			lastError = err.Error()
			continue
		}
		sent++
	}
	if sent > 0 {
		message.Status = StatusSent
		message.LastError = ""
	} else if lastError == "provider_disabled" {
		message.Status = StatusSkipped
		message.LastError = lastError
	} else {
		message.Status = StatusFailed
		message.LastError = lastError
	}
	return s.store.SaveMessage(message)
}

func validRole(role Role) bool { return role == RoleRider || role == RoleDriver }
func validPlatform(platform Platform) bool {
	return platform == PlatformAndroid || platform == PlatformIOS
}

func renderMessage(role Role, eventType string) (string, string, bool) {
	if eventType == "dispatch.offer" && role == RoleDriver {
		return "FlashX · Yêu cầu mới", "Có một công việc phù hợp đang chờ bạn phản hồi.", true
	}
	bodies := map[string]string{
		"trip.accepted":                     "Tài xế đã nhận công việc.",
		"trip.arriving":                     "Tài xế đang di chuyển tới điểm nhận.",
		"trip.arriving_for_pickup":          "Tài xế đang di chuyển tới điểm nhận.",
		"trip.arrived":                      "Tài xế đã tới điểm nhận.",
		"trip.arrived_for_pickup":           "Tài xế đã tới điểm nhận.",
		"trip.vehicle_received":             "Phương tiện đã được xác nhận bàn giao.",
		"trip.in_progress":                  "Công việc đã bắt đầu.",
		"trip.en_route_to_inspection":       "Xe đang được đưa tới nơi đăng kiểm.",
		"trip.arrived_at_inspection_center": "Xe đã tới nơi đăng kiểm.",
		"trip.inspection_in_progress":       "Đăng kiểm đang được thực hiện.",
		"trip.inspection_completed":         "Đăng kiểm đã có kết quả.",
		"trip.returning_vehicle":            "Xe đang được đưa trở lại điểm bàn giao.",
		"trip.arrived_for_return":           "Tài xế đã tới điểm trả xe.",
		"trip.handover":                     "Công việc đang ở bước bàn giao.",
		"trip.completed":                    "Công việc đã hoàn thành.",
		"trip.cancelled":                    "Công việc đã được hủy.",
		"trip.incident":                     "Công việc có sự cố cần được theo dõi.",
		"trip.searching":                    "FlashX đang tìm tài xế phù hợp.",
	}
	body, ok := bodies[eventType]
	if !ok {
		return "", "", false
	}
	if role == RoleDriver && eventType == "trip.accepted" {
		body = "Bạn đã nhận công việc. Hãy kiểm tra điểm nhận và thông tin phương tiện."
	}
	return "FlashX · Cập nhật công việc", body, true
}
