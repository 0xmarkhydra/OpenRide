package driverdocs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"flashx/services/api/internal/objectstorage"
	"flashx/services/api/internal/platform/ids"
)

const maxDocumentSize = 15 * 1024 * 1024

var (
	ErrInvalidInput       = errors.New("invalid driver document input")
	ErrForbidden          = errors.New("driver document forbidden")
	ErrStorageUnavailable = errors.New("driver document storage unavailable")
)

var allowedDocumentTypes = map[string]struct{}{
	"identity_front":       {},
	"identity_back":        {},
	"driver_license":       {},
	"vehicle_registration": {},
	"vehicle_insurance":    {},
	"portrait":             {},
}

var allowedContentTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"application/pdf": {},
}

type Service struct {
	store  Store
	signer objectstorage.Signer
	now    func() time.Time
}

type UploadTicket struct {
	DocumentID   string                      `json:"document_id"`
	DocumentType string                      `json:"document_type"`
	ObjectKey    string                      `json:"object_key"`
	Upload       objectstorage.SignedRequest `json:"upload"`
}

type DocumentWithURL struct {
	Document Document                    `json:"document"`
	View     objectstorage.SignedRequest `json:"view"`
}

func NewService(store Store, signer objectstorage.Signer) *Service {
	return &Service{
		store:  store,
		signer: signer,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) PrepareUpload(ctx context.Context, driverID string, input UploadInput) (UploadTicket, error) {
	if s.signer == nil {
		return UploadTicket{}, ErrStorageUnavailable
	}
	input = normalizeUploadInput(input)
	if !validUploadInput(input) || strings.TrimSpace(driverID) == "" {
		return UploadTicket{}, ErrInvalidInput
	}
	documentID := ids.New("doc")
	key := buildObjectKey(driverID, input.DocumentType, documentID, input.Filename, s.now())
	signed, err := s.signer.SignPut(ctx, key, input.ContentType, input.SizeBytes)
	if err != nil {
		return UploadTicket{}, fmt.Errorf("%w: %v", ErrStorageUnavailable, err)
	}
	return UploadTicket{
		DocumentID:   documentID,
		DocumentType: input.DocumentType,
		ObjectKey:    key,
		Upload:       signed,
	}, nil
}

func (s *Service) CompleteUpload(driverID string, input CompleteInput) (Document, error) {
	input.DocumentID = strings.TrimSpace(input.DocumentID)
	input.DocumentType = strings.TrimSpace(strings.ToLower(input.DocumentType))
	input.ObjectKey = strings.TrimSpace(input.ObjectKey)
	input.Filename = strings.TrimSpace(input.Filename)
	input.ContentType = strings.TrimSpace(strings.ToLower(input.ContentType))
	if input.DocumentID == "" || input.ObjectKey == "" || strings.TrimSpace(driverID) == "" ||
		!validUploadInput(UploadInput{DocumentType: input.DocumentType, Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes}) {
		return Document{}, ErrInvalidInput
	}
	if !objectKeyBelongsTo(driverID, input.DocumentType, input.DocumentID, input.ObjectKey) {
		return Document{}, ErrForbidden
	}
	now := s.now()
	doc := Document{
		ID:           input.DocumentID,
		DriverID:     driverID,
		DocumentType: input.DocumentType,
		ObjectKey:    input.ObjectKey,
		Filename:     input.Filename,
		ContentType:  input.ContentType,
		SizeBytes:    input.SizeBytes,
		ReviewStatus: ReviewPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.Create(doc); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *Service) ListForDriver(driverID string) ([]Document, error) {
	return s.store.ListForDriver(driverID)
}

func (s *Service) ViewForDriver(ctx context.Context, driverID, documentID string) (DocumentWithURL, error) {
	doc, err := s.store.Get(documentID)
	if err != nil {
		return DocumentWithURL{}, err
	}
	if doc.DriverID != driverID {
		return DocumentWithURL{}, ErrForbidden
	}
	return s.signView(ctx, doc)
}

func (s *Service) ListForAdmin(ctx context.Context, driverID string) ([]DocumentWithURL, error) {
	docs, err := s.store.ListForDriver(driverID)
	if err != nil {
		return nil, err
	}
	result := make([]DocumentWithURL, 0, len(docs))
	for _, doc := range docs {
		item, err := s.signView(ctx, doc)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) Review(driverID, documentID string, status ReviewStatus, note string) (Document, error) {
	if status != ReviewApproved && status != ReviewRejected {
		return Document{}, ErrInvalidInput
	}
	doc, err := s.store.Get(documentID)
	if err != nil {
		return Document{}, err
	}
	if doc.DriverID != driverID {
		return Document{}, ErrForbidden
	}
	return s.store.Review(documentID, status, strings.TrimSpace(note))
}

func (s *Service) signView(ctx context.Context, doc Document) (DocumentWithURL, error) {
	if s.signer == nil {
		return DocumentWithURL{}, ErrStorageUnavailable
	}
	signed, err := s.signer.SignGet(ctx, doc.ObjectKey)
	if err != nil {
		return DocumentWithURL{}, fmt.Errorf("%w: %v", ErrStorageUnavailable, err)
	}
	return DocumentWithURL{Document: doc, View: signed}, nil
}

func normalizeUploadInput(input UploadInput) UploadInput {
	input.DocumentType = strings.TrimSpace(strings.ToLower(input.DocumentType))
	input.Filename = strings.TrimSpace(input.Filename)
	input.ContentType = strings.TrimSpace(strings.ToLower(input.ContentType))
	return input
}

func validUploadInput(input UploadInput) bool {
	if _, ok := allowedDocumentTypes[input.DocumentType]; !ok {
		return false
	}
	if _, ok := allowedContentTypes[input.ContentType]; !ok {
		return false
	}
	return input.Filename != "" && len(input.Filename) <= 180 && input.SizeBytes > 0 && input.SizeBytes <= maxDocumentSize
}

func buildObjectKey(driverID, documentType, documentID, filename string, now time.Time) string {
	return fmt.Sprintf(
		"flashx/kyc/drivers/%s/%s/%04d/%02d/%s-%s",
		safeSegment(driverID), documentType, now.Year(), int(now.Month()), documentID, safeFilename(filename),
	)
}

func objectKeyBelongsTo(driverID, documentType, documentID, objectKey string) bool {
	prefix := fmt.Sprintf("flashx/kyc/drivers/%s/%s/", safeSegment(driverID), documentType)
	if !strings.HasPrefix(objectKey, prefix) {
		return false
	}
	return strings.Contains(filepath.Base(objectKey), documentID+"-")
}

func safeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.TrimSuffix(name, filepath.Ext(name))
	base = strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			return r
		case unicode.IsSpace(r):
			return '-'
		default:
			return -1
		}
	}, base)
	base = strings.Trim(base, "-_")
	if base == "" {
		base = "document"
	}
	if len(base) > 80 {
		base = base[:80]
	}
	if len(ext) > 10 {
		ext = ""
	}
	return base + ext
}

func safeSegment(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, strings.TrimSpace(value))
	return strings.Trim(value, "-")
}
