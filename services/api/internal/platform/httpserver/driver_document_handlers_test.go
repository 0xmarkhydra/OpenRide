package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"flashx/services/api/internal/driverdocs"
	"flashx/services/api/internal/objectstorage"
)

type httpTestSigner struct{}

func (httpTestSigner) SignPut(_ context.Context, key, contentType string, _ int64) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{
		Method: "PUT",
		URL: "https://objects.test/upload?signature=test",
		Headers: map[string][]string{"Content-Type": []string{contentType}},
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}, nil
}

func (httpTestSigner) SignGet(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "GET", URL: "https://objects.test/" + key + "?signature=test"}, nil
}

func (httpTestSigner) SignDelete(_ context.Context, key string) (objectstorage.SignedRequest, error) {
	return objectstorage.SignedRequest{Method: "DELETE", URL: "https://objects.test/" + key + "?signature=test"}, nil
}

func TestDriverDocumentDirectUploadFlow(t *testing.T) {
	s := newTestServer()
	s.deps.DriverDocuments = driverdocs.NewService(driverdocs.NewMemoryStore(), httpTestSigner{})
	if _, err := s.deps.Drivers.RegisterApproved("driver-doc-1", "Driver Docs", "bike"); err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{"X-Dev-Driver-ID": "driver-doc-1"}
	prepare := perform(t, s, http.MethodPost, "/v1/driver/documents/upload-url", []byte(`{
		"document_type":"driver_license",
		"filename":"gplx.jpg",
		"content_type":"image/jpeg",
		"size_bytes":2048
	}`), headers)
	if prepare.Code != http.StatusOK {
		t.Fatalf("prepare status=%d body=%s", prepare.Code, prepare.Body.String())
	}
	var prepared struct {
		Data driverdocs.UploadTicket `json:"data"`
	}
	if err := json.Unmarshal(prepare.Body.Bytes(), &prepared); err != nil {
		t.Fatal(err)
	}
	if prepared.Data.Upload.Method != "PUT" || prepared.Data.Upload.URL == "" {
		t.Fatalf("unexpected upload ticket %#v", prepared.Data)
	}

	completeBody, err := json.Marshal(driverdocs.CompleteInput{
		DocumentID: prepared.Data.DocumentID,
		DocumentType: prepared.Data.DocumentType,
		ObjectKey: prepared.Data.ObjectKey,
		Filename: "gplx.jpg",
		ContentType: "image/jpeg",
		SizeBytes: 2048,
	})
	if err != nil {
		t.Fatal(err)
	}
	complete := perform(t, s, http.MethodPost, "/v1/driver/documents/complete", completeBody, headers)
	if complete.Code != http.StatusCreated {
		t.Fatalf("complete status=%d body=%s", complete.Code, complete.Body.String())
	}

	list := perform(t, s, http.MethodGet, "/v1/driver/documents", nil, headers)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
}
