package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrUnavailable   = errors.New("object storage unavailable")
	ErrInvalidConfig = errors.New("invalid object storage config")
)

type SignedRequest struct {
	Method    string              `json:"method"`
	URL       string              `json:"url"`
	Headers   map[string][]string `json:"headers,omitempty"`
	ExpiresAt time.Time           `json:"expires_at"`
}

type Signer interface {
	SignPut(ctx context.Context, key, contentType string, contentLength int64) (SignedRequest, error)
	SignGet(ctx context.Context, key string) (SignedRequest, error)
	SignDelete(ctx context.Context, key string) (SignedRequest, error)
}

type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
	TTL             time.Duration
}

type S3Signer struct {
	bucket    string
	presigner *s3.PresignClient
	ttl       time.Duration
	now       func() time.Time
}

func NewS3Signer(ctx context.Context, cfg S3Config) (*S3Signer, error) {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.AccessKeyID = strings.TrimSpace(cfg.AccessKeyID)
	cfg.SecretAccessKey = strings.TrimSpace(cfg.SecretAccessKey)
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("%w: endpoint, region, bucket and credentials are required", ErrInvalidConfig)
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = cfg.ForcePathStyle
	})
	return &S3Signer{
		bucket:    cfg.Bucket,
		presigner: s3.NewPresignClient(client),
		ttl:       cfg.TTL,
		now:       func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *S3Signer) SignPut(ctx context.Context, key, contentType string, contentLength int64) (SignedRequest, error) {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(contentType) == "" || contentLength <= 0 {
		return SignedRequest{}, ErrInvalidConfig
	}
	request, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}, func(options *s3.PresignOptions) {
		options.Expires = s.ttl
	})
	if err != nil {
		return SignedRequest{}, fmt.Errorf("presign put object: %w", err)
	}
	return signedRequest(request.Method, request.URL, request.SignedHeader, s.now().Add(s.ttl)), nil
}

func (s *S3Signer) SignGet(ctx context.Context, key string) (SignedRequest, error) {
	if strings.TrimSpace(key) == "" {
		return SignedRequest{}, ErrInvalidConfig
	}
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(options *s3.PresignOptions) {
		options.Expires = s.ttl
	})
	if err != nil {
		return SignedRequest{}, fmt.Errorf("presign get object: %w", err)
	}
	return signedRequest(request.Method, request.URL, request.SignedHeader, s.now().Add(s.ttl)), nil
}

func (s *S3Signer) SignDelete(ctx context.Context, key string) (SignedRequest, error) {
	if strings.TrimSpace(key) == "" {
		return SignedRequest{}, ErrInvalidConfig
	}
	request, err := s.presigner.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(options *s3.PresignOptions) {
		options.Expires = s.ttl
	})
	if err != nil {
		return SignedRequest{}, fmt.Errorf("presign delete object: %w", err)
	}
	return signedRequest(request.Method, request.URL, request.SignedHeader, s.now().Add(s.ttl)), nil
}

func signedRequest(method, url string, headers http.Header, expiresAt time.Time) SignedRequest {
	copied := make(map[string][]string, len(headers))
	for key, values := range headers {
		copied[key] = append([]string(nil), values...)
	}
	return SignedRequest{Method: method, URL: url, Headers: copied, ExpiresAt: expiresAt}
}
