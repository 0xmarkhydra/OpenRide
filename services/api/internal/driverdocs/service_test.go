package driverdocs

import (
	"context"
	"strings"
	"testing"
	"time"

	"flashx/services/api/internal/objectstorage"
)

type fakeSigner struct {
	lastPutKey      string
	lastContentType string
	lastLength      int64
}

func (f *fakeSigner) SignPut(_ context.Context, key, contentType string, contentLength int64) (objectstorage.SignedRequest, error) {
	f.lastPutKey = key
	f.lastContentType = contentType
	f.lastLength = contentLength
	return objectstorage.SignedRequest{
		Method:    "PUT",
		URL:       "https://storage.test/upload",
		Headers:   map[string][]string{"Content-Type": []string{contentType}},
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}, nil
}

func (f *fakeSigner) SignGet(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "GET", URL: "https://storage.test/" + key}, nil
}

func (f *fakeSigner) SignDelete(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "DELETE", URL: "https://storage.test/" + key}, nil
}

func TestPrepareUploadReturnsDirectStorageRequest(t *testing.T) {
	signer := &fakeSigner{}
	service := NewService(NewMemoryStore(), signer)
	ticket, err := service.PrepareUpload(context.Background(), "drv_123", UploadInput{
		DocumentType: "driver_license",
		Filename:     "GPLX mat truoc.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    1024,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Upload.Method != "PUT" || !strings.HasPrefix(ticket.Upload.URL, "https://storage.test/") {
		t.Fatalf("expected direct storage PUT request, got %#v", ticket.Upload)
	}
	if signer.lastPutKey != ticket.ObjectKey || signer.lastContentType != "image/jpeg" || signer.lastLength != 1024 {
		t.Fatalf("unexpected signer input: %#v", signer)
	}
	if !strings.HasPrefix(ticket.ObjectKey, "flashx/kyc/drivers/drv_123/driver_license/") {
		t.Fatalf("unexpected key %q", ticket.ObjectKey)
	}
}

func TestCompleteUploadRejectsForeignNamespace(t *testing.T) {
	service := NewService(NewMemoryStore(), &fakeSigner{})
	_, err := service.CompleteUpload("drv_a", CompleteInput{
		DocumentID:   "doc_123",
		DocumentType: "portrait",
		ObjectKey:    "flashx/kyc/drivers/drv_b/portrait/2026/08/doc_123-photo.jpg",
		Filename:     "photo.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    1024,
	})
	if err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestCompleteListViewAndReview(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, &fakeSigner{})
	service.now = func() time.Time { return time.Date(2026, 8, 11, 1, 0, 0, 0, time.UTC) }

	ticket, err := service.PrepareUpload(context.Background(), "drv_1", UploadInput{
		DocumentType: "identity_front",
		Filename:     "cccd.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := service.CompleteUpload("drv_1", CompleteInput{
		DocumentID:   ticket.DocumentID,
		DocumentType: ticket.DocumentType,
		ObjectKey:    ticket.ObjectKey,
		Filename:     "cccd.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := service.ListForDriver("drv_1")
	if err != nil || len(items) != 1 {
		t.Fatalf("list=%v err=%v", items, err)
	}
	view, err := service.ViewForDriver(context.Background(), "drv_1", doc.ID)
	if err != nil || view.View.Method != "GET" {
		t.Fatalf("view=%#v err=%v", view, err)
	}
	reviewed, err := service.Review("drv_1", doc.ID, ReviewApproved, "verified")
	if err != nil || reviewed.ReviewStatus != ReviewApproved {
		t.Fatalf("reviewed=%#v err=%v", reviewed, err)
	}
}

func TestPrepareUploadValidatesTypeAndSize(t *testing.T) {
	service := NewService(NewMemoryStore(), &fakeSigner{})
	_, err := service.PrepareUpload(context.Background(), "drv_1", UploadInput{
		DocumentType: "video",
		Filename:     "x.mp4",
		ContentType:  "video/mp4",
		SizeBytes:    100,
	})
	if err != ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
