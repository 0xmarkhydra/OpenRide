package inspectionchecklist

import (
	"testing"

	"flashx/services/api/internal/trips"
)

func setupInspectionChecklist(t *testing.T) (*Service, *trips.Service, trips.Trip) {
	t.Helper()
	tripService := trips.NewService(trips.NewMemoryStore())
	trip, err := tripService.Create(trips.CreateInput{
		RiderID: "rider-inspection", ServiceType: trips.ServiceVehicleInspection,
		Pickup: trips.Point{Lat: 19.807, Lng: 105.776}, Destination: trips.Point{Lat: 19.82, Lng: 105.79},
		EstimatedDistanceM: 5000, EstimatedDurationS: 900,
		FareBreakdown: trips.FareBreakdown{TotalMinor: 299000},
	})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(NewMemoryStore(), tripService)
	if err != nil {
		t.Fatal(err)
	}
	return service, tripService, trip
}

func TestChecklistSnapshotAndReadiness(t *testing.T) {
	service, tripService, trip := setupInspectionChecklist(t)
	snapshot, err := service.ForRider(trip.ID, "rider-inspection")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Checklist.TemplateVersion != 1 || len(snapshot.Items) != 4 || snapshot.Ready {
		t.Fatalf("unexpected initial snapshot: %+v", snapshot)
	}

	if _, err := service.ReplaceTemplate(ReplaceTemplateInput{
		Reason: "Test next template",
		Items: []TemplateItem{
			{Key: "registration_v2", Label: "Giấy tờ xe V2", Required: true, SortOrder: 10},
		},
	}, "admin-1"); err != nil {
		t.Fatal(err)
	}
	originalAgain, err := service.ForRider(trip.ID, "rider-inspection")
	if err != nil {
		t.Fatal(err)
	}
	if originalAgain.Checklist.TemplateVersion != 1 || originalAgain.Items[0].Key == "registration_v2" {
		t.Fatalf("existing job must retain old snapshot: %+v", originalAgain)
	}

	var requiredKey string
	for _, item := range snapshot.Items {
		status := CustomerNotApplicable
		if item.Required {
			requiredKey = item.Key
			status = CustomerPresent
		}
		if _, err := service.UpdateCustomer(trip.ID, "rider-inspection", item.Key, CustomerUpdate{Status: status}); err != nil {
			t.Fatalf("customer update %s: %v", item.Key, err)
		}
	}

	if _, err := tripService.AssignDriver(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArriving(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArrived(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	current, err := service.ForDriver(trip.ID, "driver-inspection")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range current.Items {
		status := DriverNotApplicable
		if item.Required {
			status = DriverReceived
		}
		if _, err := service.UpdateDriver(trip.ID, "driver-inspection", item.Key, DriverUpdate{Status: status}); err != nil {
			t.Fatalf("driver update %s: %v", item.Key, err)
		}
	}
	ready, err := service.IsReady(trip.ID)
	if err != nil || !ready {
		t.Fatalf("checklist ready=%v err=%v", ready, err)
	}

	// Customer changing the required declaration invalidates the driver's old
	// verification for that item.
	// Move trip back is intentionally impossible, so exercise the invariant on
	// a second job still before custody.
	trip2, err := tripService.Create(trips.CreateInput{
		RiderID: "rider-inspection", ServiceType: trips.ServiceVehicleInspection,
		Pickup: trips.Point{Lat: 19.807, Lng: 105.776}, Destination: trips.Point{Lat: 19.82, Lng: 105.79},
		FareBreakdown: trips.FareBreakdown{TotalMinor: 299000},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.ForRider(trip2.ID, "rider-inspection")
	if err != nil {
		t.Fatal(err)
	}
	if second.Checklist.TemplateVersion != 2 || second.Items[0].Key != "registration_v2" {
		t.Fatalf("new job should snapshot v2: %+v", second)
	}
	if _, err := service.UpdateCustomer(trip2.ID, "rider-inspection", "registration_v2", CustomerUpdate{Status: CustomerNotApplicable}); err != ErrInvalidInput {
		t.Fatalf("required customer not_applicable err=%v want ErrInvalidInput", err)
	}
	_ = requiredKey
}

func TestCustomerChangeResetsDriverVerification(t *testing.T) {
	service, tripService, trip := setupInspectionChecklist(t)
	snapshot, err := service.ForRider(trip.ID, "rider-inspection")
	if err != nil {
		t.Fatal(err)
	}
	var required Item
	for _, item := range snapshot.Items {
		if item.Required {
			required = item
			break
		}
	}
	if _, err := service.UpdateCustomer(trip.ID, "rider-inspection", required.Key, CustomerUpdate{Status: CustomerPresent, Note: "Bản gốc"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.AssignDriver(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArriving(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArrived(trip.ID, "driver-inspection"); err != nil {
		t.Fatal(err)
	}
	verified, err := service.UpdateDriver(trip.ID, "driver-inspection", required.Key, DriverUpdate{Status: DriverReceived})
	if err != nil {
		t.Fatal(err)
	}
	item, _ := findItem(verified.Items, required.Key)
	if item.DriverStatus != DriverReceived {
		t.Fatalf("driver verification missing: %+v", item)
	}

	// Simulate a customer-side declaration change at the service level by using
	// a new inspection job; customer edits are intentionally locked once driver
	// has reached pickup on the first job. The reset behavior itself is covered
	// by SaveItem mutation before pickup below.
	trip3, err := tripService.Create(trips.CreateInput{
		RiderID: "rider-inspection", ServiceType: trips.ServiceVehicleInspection,
		Pickup: trips.Point{Lat: 19.80, Lng: 105.77}, Destination: trips.Point{Lat: 19.81, Lng: 105.78},
		FareBreakdown: trips.FareBreakdown{TotalMinor: 299000},
	})
	if err != nil {
		t.Fatal(err)
	}
	third, err := service.ForRider(trip3.ID, "rider-inspection")
	if err != nil {
		t.Fatal(err)
	}
	key := third.Items[0].Key
	firstUpdate, err := service.UpdateCustomer(trip3.ID, "rider-inspection", key, CustomerUpdate{Status: CustomerPresent, Note: "Có giấy"})
	if err != nil {
		t.Fatal(err)
	}
	item, _ = findItem(firstUpdate.Items, key)
	item.DriverStatus = DriverReceived
	if err := service.store.SaveItem(item); err != nil {
		t.Fatal(err)
	}
	changed, err := service.UpdateCustomer(trip3.ID, "rider-inspection", key, CustomerUpdate{Status: CustomerPresent, Note: "Đổi sang bản sao có chứng thực"})
	if err != nil {
		t.Fatal(err)
	}
	item, _ = findItem(changed.Items, key)
	if item.DriverStatus != DriverPending || item.DriverUpdatedAt != nil {
		t.Fatalf("customer change must reset driver verification: %+v", item)
	}
}

func TestChecklistRejectsNonInspectionTrip(t *testing.T) {
	tripService := trips.NewService(trips.NewMemoryStore())
	trip, err := tripService.Create(trips.CreateInput{RiderID: "rider-car", ServiceType: trips.ServiceDesignatedDriverCar, Pickup: trips.Point{Lat: 19.8, Lng: 105.7}, Destination: trips.Point{Lat: 19.9, Lng: 105.8}, FareBreakdown: trips.FareBreakdown{TotalMinor: 100000}})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(NewMemoryStore(), tripService)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ForRider(trip.ID, "rider-car"); err != ErrInvalidInput {
		t.Fatalf("non-inspection err=%v want ErrInvalidInput", err)
	}
}
