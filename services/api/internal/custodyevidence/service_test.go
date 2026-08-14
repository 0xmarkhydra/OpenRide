package custodyevidence

import (
	"context"
	"testing"
	"time"

	"flashx/services/api/internal/objectstorage"
	"flashx/services/api/internal/trips"
)

type fakeSigner struct{}

func (fakeSigner) SignPut(_ context.Context, key, contentType string, contentLength int64) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "PUT", URL: "https://storage.test/put/" + key, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (fakeSigner) SignGet(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "GET", URL: "https://storage.test/get/" + key, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (fakeSigner) SignDelete(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "DELETE", URL: "https://storage.test/delete/" + key, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func setupEvidenceTrip(t *testing.T) (*Service, *trips.Service, trips.Trip) {
	t.Helper()
	tripService := trips.NewService(trips.NewMemoryStore())
	trip, err := tripService.Create(trips.CreateInput{
		RiderID:            "rider-evidence",
		ServiceType:        trips.ServiceDesignatedDriverCar,
		BookingMode:        trips.BookingImmediate,
		Pickup:             trips.Point{Lat: 19.807, Lng: 105.776},
		Destination:        trips.Point{Lat: 19.82, Lng: 105.79},
		FareBreakdown:      trips.FareBreakdown{TotalMinor: 150000},
		EstimatedDistanceM: 5000,
		EstimatedDurationS: 900,
	})
	if err != nil {
		t.Fatal(err)
	}
	if trip, err = tripService.AssignDriver(trip.ID, "driver-evidence"); err != nil {
		t.Fatal(err)
	}
	if trip, err = tripService.MarkArriving(trip.ID, "driver-evidence"); err != nil {
		t.Fatal(err)
	}
	if trip, err = tripService.MarkArrived(trip.ID, "driver-evidence"); err != nil {
		t.Fatal(err)
	}
	return NewService(NewMemoryStore(), fakeSigner{}, tripService), tripService, trip
}

func TestPickupEvidenceRequiresPhotosAndBilateralConfirmation(t *testing.T) {
	service, _, trip := setupEvidenceTrip(t)
	odometer := int64(18342)
	fuel := 72
	snapshot, err := service.UpdateByDriver(trip.ID, "driver-evidence", StagePickup, EvidenceInput{
		ConditionNote: "Không thấy vết móp mới; ngoại thất khô ráo.",
		OdometerKm:    &odometer,
		FuelPercent:   &fuel,
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Ready || snapshot.Evidence.DriverConfirmedAt != nil || snapshot.Evidence.RiderConfirmedAt != nil {
		t.Fatalf("new metadata must not be ready: %+v", snapshot)
	}

	for index, photoType := range []string{"front", "rear"} {
		ticket, err := service.PreparePhotoUpload(context.Background(), trip.ID, "driver-evidence", StagePickup, PhotoUploadInput{
			PhotoType: photoType, Filename: photoType + ".jpg", ContentType: "image/jpeg", SizeBytes: 1024 + int64(index),
		})
		if err != nil {
			t.Fatal(err)
		}
		snapshot, err = service.CompletePhotoUpload(trip.ID, "driver-evidence", StagePickup, CompletePhotoInput{
			PhotoID: ticket.PhotoID, PhotoType: ticket.PhotoType, ObjectKey: ticket.ObjectKey,
			Filename: photoType + ".jpg", ContentType: "image/jpeg", SizeBytes: 1024 + int64(index),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(snapshot.Photos) != 2 || snapshot.Ready {
		t.Fatalf("two photos without confirmation must not be ready: %+v", snapshot)
	}

	snapshot, err = service.ConfirmDriver(trip.ID, "driver-evidence", StagePickup)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Evidence.DriverConfirmedAt == nil || snapshot.Ready {
		t.Fatalf("driver confirmation alone must not be ready: %+v", snapshot)
	}

	snapshot, err = service.ConfirmRider(trip.ID, "rider-evidence", StagePickup)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Ready || snapshot.Evidence.RiderConfirmedAt == nil {
		t.Fatalf("bilateral confirmed evidence must be ready: %+v", snapshot)
	}
	if snapshot.Photos[0].View == nil || snapshot.Photos[0].View.URL == "" {
		t.Fatalf("signed photo view missing: %+v", snapshot.Photos[0])
	}

	snapshot, err = service.UpdateByDriver(trip.ID, "driver-evidence", StagePickup, EvidenceInput{
		ConditionNote: "Bổ sung ghi chú: có vết xước cũ ở cản sau.",
		OdometerKm:    &odometer,
		FuelPercent:   &fuel,
	})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Ready || snapshot.Evidence.DriverConfirmedAt != nil || snapshot.Evidence.RiderConfirmedAt != nil {
		t.Fatalf("mutating confirmed evidence must reset confirmations: %+v", snapshot)
	}
}

func TestEvidencePermissionsStateAndStorage(t *testing.T) {
	service, tripService, trip := setupEvidenceTrip(t)
	if _, err := service.UpdateByDriver(trip.ID, "other-driver", StagePickup, EvidenceInput{ConditionNote: "Xe bình thường"}); err != ErrForbidden {
		t.Fatalf("wrong driver err=%v want ErrForbidden", err)
	}
	if _, err := service.UpdateByDriver(trip.ID, "driver-evidence", StageReturn, EvidenceInput{ConditionNote: "Chưa tới giai đoạn trả xe"}); err != ErrInvalidState {
		t.Fatalf("return before service start err=%v want ErrInvalidState", err)
	}
	if _, err := service.ConfirmRider(trip.ID, "other-rider", StagePickup); err != ErrForbidden {
		t.Fatalf("wrong rider err=%v want ErrForbidden", err)
	}

	withoutStorage := NewService(NewMemoryStore(), nil, tripService)
	if _, err := withoutStorage.UpdateByDriver(trip.ID, "driver-evidence", StagePickup, EvidenceInput{ConditionNote: "Xe bình thường"}); err != nil {
		t.Fatal(err)
	}
	if _, err := withoutStorage.PreparePhotoUpload(context.Background(), trip.ID, "driver-evidence", StagePickup, PhotoUploadInput{
		PhotoType: "front", Filename: "front.jpg", ContentType: "image/jpeg", SizeBytes: 1024,
	}); err != ErrStorageUnavailable {
		t.Fatalf("prepare without storage err=%v want ErrStorageUnavailable", err)
	}

	if _, err := tripService.MarkVehicleReceived(trip.ID, "driver-evidence"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.Start(trip.ID, "driver-evidence"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateByDriver(trip.ID, "driver-evidence", StageReturn, EvidenceInput{ConditionNote: "Xe sẵn sàng bàn giao"}); err != nil {
		t.Fatalf("return evidence during in-progress should be allowed: %v", err)
	}
}

func TestConfirmNeedsTwoPhotos(t *testing.T) {
	service, _, trip := setupEvidenceTrip(t)
	if _, err := service.UpdateByDriver(trip.ID, "driver-evidence", StagePickup, EvidenceInput{ConditionNote: "Xe bình thường"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmDriver(trip.ID, "driver-evidence", StagePickup); err != ErrNotReady {
		t.Fatalf("confirm without photos err=%v want ErrNotReady", err)
	}
}
