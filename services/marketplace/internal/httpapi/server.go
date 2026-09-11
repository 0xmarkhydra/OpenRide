package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/engine"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/ranking"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

const maxPublicMoneyMinor int64 = 9_007_199_254_740_991

type Server struct {
	http     *http.Server
	services *extension.Registry
	ready    func(context.Context) error
	v2       V2Store
	accept   app.AcceptanceService
	ranker   engine.Ranker
}

type Config struct {
	Addr       string
	Services   *extension.Registry
	Ready      func(context.Context) error
	V2Store    V2Store
	Acceptance app.AcceptanceService
	Ranker     engine.Ranker
}

type envelope struct { Data any `json:"data"` }
type apiError struct { Code string `json:"code"`; Message string `json:"message"` }
type errorEnvelope struct { Error apiError `json:"error"` }

func New(cfg Config) (*Server, error) {
	if cfg.Services == nil { return nil, errors.New("marketplace http: service registry is required") }
	if cfg.Addr == "" { cfg.Addr = ":8090" }
	if cfg.V2Store != nil && cfg.Acceptance.Store == nil { return nil, errors.New("marketplace http: acceptance service is required when V2 store is enabled") }
	if cfg.Ranker == nil { cfg.Ranker = ranking.Default() }

	s := &Server{services: cfg.Services, ready: cfg.Ready, v2: adaptV2Store(cfg.V2Store), accept: cfg.Acceptance, ranker: cfg.Ranker}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.readiness)
	mux.HandleFunc("GET /internal/v1/services", s.listServices)
	mux.HandleFunc("POST /internal/v1/requests/validate", s.validateRequest)
	s.registerV2(mux)

	s.http = &http.Server{Addr: cfg.Addr, Handler: middleware(mux), ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second, WriteTimeout: 15*time.Second, IdleTimeout: 60*time.Second}
	return s,nil
}

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }
func (s *Server) Handler() http.Handler { return s.http.Handler }

func (s *Server) health(w http.ResponseWriter, _ *http.Request) { writeJSON(w,http.StatusOK,envelope{Data:map[string]any{"status":"ok","service":"marketplace-service"}}) }
func (s *Server) readiness(w http.ResponseWriter,r *http.Request) { if s.ready!=nil { ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();if err:=s.ready(ctx);err!=nil{writeError(w,http.StatusServiceUnavailable,"NOT_READY","marketplace dependency unavailable");return} }; writeJSON(w,http.StatusOK,envelope{Data:map[string]any{"status":"ready"}}) }
func (s *Server) listServices(w http.ResponseWriter,_ *http.Request) { writeJSON(w,http.StatusOK,envelope{Data:s.services.Manifests()}) }

func (s *Server) validateRequest(w http.ResponseWriter,r *http.Request) {
	var request marketplace.Request;if !decodeJSON(w,r,&request){return}
	if err:=request.Validate();err!=nil{writeError(w,http.StatusUnprocessableEntity,"REQUEST_INVALID",err.Error());return}
	module,err:=s.services.Get(request.ServiceType);if err!=nil{writeError(w,http.StatusUnprocessableEntity,"SERVICE_TYPE_UNSUPPORTED",err.Error());return}
	if err:=module.ValidateRequest(r.Context(),request);err!=nil{writeError(w,http.StatusUnprocessableEntity,"SERVICE_REQUEST_INVALID",err.Error());return}
	writeJSON(w,http.StatusOK,envelope{Data:map[string]any{"valid":true,"service_type":request.ServiceType,"module":module.Manifest()}})
}

func canonicalRequestHash(r *http.Request) (string, error) {
	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
		if err != nil { return "", err }
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	canonical := body
	if len(bytes.TrimSpace(body)) > 0 {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()
		var value any
		if err := dec.Decode(&value); err == nil {
			if err := dec.Decode(&struct{}{}); err == io.EOF {
				if normalized, err := json.Marshal(value); err == nil { canonical = normalized }
			}
		}
	}
	sum := sha256.Sum256(append([]byte(r.Method+"\n"+r.URL.Path+"\n"), canonical...))
	return hex.EncodeToString(sum[:]), nil
}

type bufferedResponseWriter struct {
	header http.Header
	status int
	body bytes.Buffer
}
func newBufferedResponseWriter() *bufferedResponseWriter { return &bufferedResponseWriter{header: make(http.Header)} }
func (w *bufferedResponseWriter) Header() http.Header { return w.header }
func (w *bufferedResponseWriter) WriteHeader(status int) { if w.status==0 { w.status=status } }
func (w *bufferedResponseWriter) Write(p []byte) (int,error) { if w.status==0 { w.status=http.StatusOK }; return w.body.Write(p) }
func (w *bufferedResponseWriter) flush(dst http.ResponseWriter) {
	for key, values := range w.header { for _, value := range values { dst.Header().Add(key,value) } }
	status:=w.status;if status==0 { status=http.StatusOK }
	dst.WriteHeader(status)
	_,_=dst.Write(w.body.Bytes())
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		key:=strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if key=="" {
			w.Header().Set("X-Content-Type-Options","nosniff")
			w.Header().Set("Cache-Control","no-store")
			next.ServeHTTP(w,r)
			return
		}
		hash,err:=canonicalRequestHash(r)
		if err!=nil { writeError(w,http.StatusBadRequest,"IDEMPOTENCY_HASH_FAILED","could not read request body");return }
		state:=&idempotencyReplayState{}
		ctx:=app.WithIdempotency(r.Context(),key,hash)
		r=r.WithContext(withReplayState(ctx,state))
		bw:=newBufferedResponseWriter()
		bw.Header().Set("X-Content-Type-Options","nosniff")
		bw.Header().Set("Cache-Control","no-store")
		next.ServeHTTP(bw,r)
		if state.Replay!=nil {
			bw.body.Reset();bw.status=0
			writeJSON(bw,state.Replay.Status,envelope{Data:state.Replay.Data})
		}
		bw.flush(w)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return false
	}

	genericDecoder := json.NewDecoder(bytes.NewReader(body))
	genericDecoder.UseNumber()
	var generic any
	if err := genericDecoder.Decode(&generic); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return false
	}
	if err := genericDecoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "request body must contain exactly one JSON value")
		return false
	}
	if err := validatePublicMinorFields(generic); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "MONEY_MINOR_INVALID", err.Error())
		return false
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "request body must contain exactly one JSON value")
		return false
	}
	return true
}

func validatePublicMinorFields(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if strings.HasSuffix(strings.ToLower(key), "_minor") {
				number, ok := child.(json.Number)
				if !ok {
					return fmt.Errorf("%s must be a non-negative integer no greater than %d", key, maxPublicMoneyMinor)
				}
				minor, err := number.Int64()
				if err != nil || minor < 0 || minor > maxPublicMoneyMinor {
					return fmt.Errorf("%s must be a non-negative integer no greater than %d", key, maxPublicMoneyMinor)
				}
			}
			if err := validatePublicMinorFields(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := validatePublicMinorFields(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeError(w http.ResponseWriter,status int,code,message string){if strings.Contains(message,"idempotency key reused with different payload"){status=http.StatusConflict;code="IDEMPOTENCY_CONFLICT";message="Idempotency-Key was already used with a different request"};writeJSON(w,status,errorEnvelope{Error:apiError{Code:code,Message:message}})}
func writeJSON(w http.ResponseWriter,status int,payload any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_=json.NewEncoder(w).Encode(payload)}
