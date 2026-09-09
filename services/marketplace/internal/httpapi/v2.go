package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

type V2Store interface {
	CreateTariff(context.Context, marketplace.DriverTariff) error
	ListDriverTariffs(context.Context, string) ([]marketplace.DriverTariff, error)
	CreateRequest(context.Context, marketplace.Request) error
	GetRequest(context.Context, string) (marketplace.Request, error)
	CancelRequest(context.Context, string, string) error
	ListEligibleRequests(context.Context, string, marketplace.ServiceType) ([]marketplace.Request, error)
	SubmitQuote(context.Context, marketplace.Quote) error
	ListOffers(context.Context, string) ([]marketplace.Quote, error)
	WithdrawQuote(context.Context, string, string) error
}

func (s *Server) registerV2(mux *http.ServeMux) {
	mux.HandleFunc("GET /v2/services", s.listServices)
	if s.v2 == nil { return }
	mux.HandleFunc("POST /v2/requests", s.createRequest)
	mux.HandleFunc("GET /v2/requests/{requestID}", s.getRequest)
	mux.HandleFunc("POST /v2/requests/{requestID}/cancel", s.cancelRequest)
	mux.HandleFunc("GET /v2/requests/{requestID}/offers", s.listOffers)
	mux.HandleFunc("POST /v2/requests/{requestID}/quotes", s.submitQuote)
	mux.HandleFunc("POST /v2/quotes/{quoteID}/accept", s.acceptQuote)
	mux.HandleFunc("POST /v2/quotes/{quoteID}/withdraw", s.withdrawQuote)
	mux.HandleFunc("POST /v2/drivers/me/tariffs", s.createDriverTariff)
	mux.HandleFunc("GET /v2/drivers/me/tariffs", s.listDriverTariffs)
	mux.HandleFunc("GET /v2/drivers/me/requests", s.listEligibleRequests)
}

func actorID(r *http.Request) (string, bool) {
	id := strings.TrimSpace(r.Header.Get("X-OpenRide-Actor-ID"))
	return id, id != ""
}

func requireActor(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := actorID(r)
	if !ok { writeError(w,http.StatusUnauthorized,"ACTOR_REQUIRED","trusted gateway must provide X-OpenRide-Actor-ID"); return "",false }
	return id,true
}

func idempotencyKey(w http.ResponseWriter, r *http.Request) (string,bool) {
	key:=strings.TrimSpace(r.Header.Get("Idempotency-Key")); if key=="" { writeError(w,http.StatusBadRequest,"IDEMPOTENCY_KEY_REQUIRED","Idempotency-Key header is required"); return "",false }; return key,true
}

func newID(prefix string) (string,error) { var raw [16]byte; if _,err:=rand.Read(raw[:]);err!=nil{return "",err}; return prefix+"_"+hex.EncodeToString(raw[:]),nil }

func (s *Server) createRequest(w http.ResponseWriter,r *http.Request) {
	rider,ok:=requireActor(w,r);if !ok{return}; if _,ok=idempotencyKey(w,r);!ok{return}
	var req marketplace.Request; if !decodeJSON(w,r,&req){return}
	id,err:=newID("req");if err!=nil{writeError(w,500,"ID_GENERATION_FAILED","could not create request id");return}
	now:=time.Now().UTC(); req.ID=id;req.RiderID=rider;req.Status=marketplace.RequestOpen;req.RequestedAt=now;req.Version=1
	if req.InstanceID==""{req.InstanceID="default"}
	if err:=req.Validate();err!=nil{writeError(w,422,"REQUEST_INVALID",err.Error());return}
	module,err:=s.services.Get(req.ServiceType);if err!=nil{writeError(w,422,"SERVICE_TYPE_UNSUPPORTED",err.Error());return};if err:=module.ValidateRequest(r.Context(),req);err!=nil{writeError(w,422,"SERVICE_REQUEST_INVALID",err.Error());return}
	if err:=s.v2.CreateRequest(r.Context(),req);err!=nil{writeError(w,409,"REQUEST_CREATE_FAILED",err.Error());return};writeJSON(w,http.StatusCreated,envelope{Data:req})
}

func (s *Server) getRequest(w http.ResponseWriter,r *http.Request){ actor,ok:=requireActor(w,r);if !ok{return};req,err:=s.v2.GetRequest(r.Context(),r.PathValue("requestID"));if err!=nil{writeError(w,404,"REQUEST_NOT_FOUND","request not found");return};if req.RiderID!=actor{writeError(w,403,"REQUEST_FORBIDDEN","request belongs to another rider");return};writeJSON(w,200,envelope{Data:req}) }
func (s *Server) cancelRequest(w http.ResponseWriter,r *http.Request){ actor,ok:=requireActor(w,r);if !ok{return};if _,ok=idempotencyKey(w,r);!ok{return};if err:=s.v2.CancelRequest(r.Context(),r.PathValue("requestID"),actor);err!=nil{writeError(w,409,"REQUEST_CANCEL_FAILED",err.Error());return};writeJSON(w,200,envelope{Data:map[string]any{"status":"cancelled"}}) }
func (s *Server) listOffers(w http.ResponseWriter,r *http.Request){ actor,ok:=requireActor(w,r);if !ok{return};req,err:=s.v2.GetRequest(r.Context(),r.PathValue("requestID"));if err!=nil||req.RiderID!=actor{writeError(w,404,"REQUEST_NOT_FOUND","request not found");return};offers,err:=s.v2.ListOffers(r.Context(),req.ID);if err!=nil{writeError(w,500,"OFFERS_READ_FAILED",err.Error());return};writeJSON(w,200,envelope{Data:offers}) }

func (s *Server) createDriverTariff(w http.ResponseWriter,r *http.Request){ driver,ok:=requireActor(w,r);if !ok{return};if _,ok=idempotencyKey(w,r);!ok{return};var t marketplace.DriverTariff;if !decodeJSON(w,r,&t){return};id,err:=newID("tariff");if err!=nil{writeError(w,500,"ID_GENERATION_FAILED","could not create tariff id");return};t.ID=id;t.DriverID=driver;if t.InstanceID==""{t.InstanceID="default"};if t.Version<1{t.Version=1};if err:=t.Validate();err!=nil{writeError(w,422,"TARIFF_INVALID",err.Error());return};if err:=s.v2.CreateTariff(r.Context(),t);err!=nil{writeError(w,409,"TARIFF_CREATE_FAILED",err.Error());return};writeJSON(w,201,envelope{Data:t}) }
func (s *Server) listDriverTariffs(w http.ResponseWriter,r *http.Request){ driver,ok:=requireActor(w,r);if !ok{return};items,err:=s.v2.ListDriverTariffs(r.Context(),driver);if err!=nil{writeError(w,500,"TARIFF_READ_FAILED",err.Error());return};writeJSON(w,200,envelope{Data:items}) }
func (s *Server) listEligibleRequests(w http.ResponseWriter,r *http.Request){ _,ok:=requireActor(w,r);if !ok{return};service:=marketplace.ServiceType(strings.TrimSpace(r.URL.Query().Get("service_type")));if service==""{service="passenger.car"};items,err:=s.v2.ListEligibleRequests(r.Context(),"default",service);if err!=nil{writeError(w,500,"REQUEST_READ_FAILED",err.Error());return};writeJSON(w,200,envelope{Data:items}) }

func (s *Server) submitQuote(w http.ResponseWriter,r *http.Request){ driver,ok:=requireActor(w,r);if !ok{return};if _,ok=idempotencyKey(w,r);!ok{return};requestID:=r.PathValue("requestID");req,err:=s.v2.GetRequest(r.Context(),requestID);if err!=nil{writeError(w,404,"REQUEST_NOT_FOUND","request not found");return};if req.Status!=marketplace.RequestOpen&&req.Status!=marketplace.RequestReceivingQuotes{writeError(w,409,"REQUEST_NOT_OPEN","request is not accepting quotes");return};var q marketplace.Quote;if !decodeJSON(w,r,&q){return};id,err:=newID("quote");if err!=nil{writeError(w,500,"ID_GENERATION_FAILED","could not create quote id");return};now:=time.Now().UTC();q.ID=id;q.RequestID=requestID;q.DriverID=driver;q.Status=marketplace.QuotePending;q.CreatedAt=now;if q.ExpiresAt.IsZero(){q.ExpiresAt=now.Add(2*time.Minute)};if err:=q.Validate();err!=nil{writeError(w,422,"QUOTE_INVALID",err.Error());return};if err:=s.v2.SubmitQuote(r.Context(),q);err!=nil{writeError(w,409,"QUOTE_CREATE_FAILED",err.Error());return};writeJSON(w,201,envelope{Data:q}) }
func (s *Server) withdrawQuote(w http.ResponseWriter,r *http.Request){ driver,ok:=requireActor(w,r);if !ok{return};if _,ok=idempotencyKey(w,r);!ok{return};if err:=s.v2.WithdrawQuote(r.Context(),r.PathValue("quoteID"),driver);err!=nil{writeError(w,409,"QUOTE_WITHDRAW_FAILED",err.Error());return};writeJSON(w,200,envelope{Data:map[string]any{"status":"withdrawn"}}) }

func (s *Server) acceptQuote(w http.ResponseWriter,r *http.Request){ rider,ok:=requireActor(w,r);if !ok{return};key,ok:=idempotencyKey(w,r);if !ok{return};agreementID,err:=newID("agr");if err!=nil{writeError(w,500,"ID_GENERATION_FAILED","could not create agreement id");return};eventID,err:=newID("evt");if err!=nil{writeError(w,500,"ID_GENERATION_FAILED","could not create event id");return};agreement,err:=s.accept.AcceptQuote(r.Context(),app.AcceptQuoteCommand{QuoteID:r.PathValue("quoteID"),RiderID:rider,IdempotencyKey:key,AgreementID:agreementID,EventID:eventID,Now:time.Now().UTC()});if err!=nil{status:=http.StatusConflict;code:="QUOTE_ACCEPT_FAILED";if errors.Is(err,app.ErrRiderMismatch){status=http.StatusForbidden;code="RIDER_MISMATCH"};writeError(w,status,code,err.Error());return};writeJSON(w,200,envelope{Data:agreement}) }
