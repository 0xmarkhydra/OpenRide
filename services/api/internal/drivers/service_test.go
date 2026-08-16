package drivers

import (
	"errors"
	"testing"
	"time"
)

func TestApprovedDriverAvailabilityAndLocation(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, err := service.RegisterApproved("driver-1", "Driver One", "bike")
	if err != nil {
		t.Fatal(err)
	}
	if driver.Availability != AvailabilityOffline {
		t.Fatalf("availability = %s, want offline", driver.Availability)
	}

	driver, err = service.SetAvailability(driver.ID, AvailabilityOnline)
	if err != nil {
		t.Fatal(err)
	}
	driver, err = service.UpdateLocation(driver.ID, Location{Lat: 21.0285, Lng: 105.8542, AccuracyM: 5, CapturedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if driver.Location == nil {
		t.Fatal("location must be stored")
	}
	if got := service.Candidates("bike", time.Minute); len(got) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(got))
	}
}

func TestBusyDriverCannotChangeAvailability(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, _ := service.RegisterApproved("driver-1", "Driver One", "bike")
	_, _ = service.SetAvailability(driver.ID, AvailabilityOnline)
	_, err := service.TryMarkBusy(driver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetAvailability(driver.ID, AvailabilityOffline); !errors.Is(err, ErrDriverUnavailable) {
		t.Fatalf("error = %v, want ErrDriverUnavailable", err)
	}
}

func TestStaleLocationIsExcluded(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, _ := service.RegisterApproved("driver-1", "Driver One", "bike")
	_, _ = service.SetAvailability(driver.ID, AvailabilityOnline)
	_, err := service.UpdateLocation(driver.ID, Location{Lat: 21.0285, Lng: 105.8542, AccuracyM: 5, CapturedAt: time.Now().UTC().Add(-3 * time.Minute)})
	if !errors.Is(err, ErrInvalidLocation) {
		t.Fatalf("error = %v, want ErrInvalidLocation", err)
	}
}

func TestPendingDriverNeedsQualificationsBeforeApproval(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, created, err := service.FindOrCreatePendingByPhone("0987000001")
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if len(driver.Capabilities) != 0 {
		t.Fatalf("new pending driver capabilities=%v, want empty", driver.Capabilities)
	}
	if _, err := service.SetApproval(driver.ID, ApprovalApproved); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("approval without qualifications err=%v, want ErrInvalidInput", err)
	}

	expiry := time.Now().UTC().AddDate(2, 0, 0)
	qualified, err := service.UpdateQualifications(driver.ID, QualificationInput{
		Capabilities:   []string{"designated_driver_car"},
		LicenseClass:   "B",
		LicenseExpiry:  &expiry,
		CanDriveManual: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(qualified.Capabilities) != 1 || qualified.Capabilities[0] != "designated_driver_car" || !qualified.CanDriveManual {
		t.Fatalf("unexpected qualifications: %+v", qualified)
	}
	approved, err := service.SetApproval(driver.ID, ApprovalApproved)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Approval != ApprovalApproved {
		t.Fatalf("approval=%s want approved", approved.Approval)
	}
}

func TestExpiredLicenseIsExcludedFromCandidates(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, _, err := service.FindOrCreatePendingByPhone("0987000002")
	if err != nil {
		t.Fatal(err)
	}
	expired := time.Now().UTC().AddDate(0, 0, -1)
	if _, err := service.UpdateQualifications(driver.ID, QualificationInput{
		Capabilities:  []string{"designated_driver_car"},
		LicenseClass:  "B",
		LicenseExpiry: &expired,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetApproval(driver.ID, ApprovalApproved); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetAvailability(driver.ID, AvailabilityOnline); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateLocation(driver.ID, Location{Lat: 21.0285, Lng: 105.8542, AccuracyM: 5, CapturedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if got := service.Candidates("designated_driver_car", time.Minute); len(got) != 0 {
		t.Fatalf("expired-license candidates=%d, want 0", len(got))
	}
}
