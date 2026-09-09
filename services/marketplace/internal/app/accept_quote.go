package app

import (
	"context"
	"errors"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var (
	ErrInvalidCommand = errors.New("marketplace app: invalid command")
	ErrRequestUnavailable = errors.New("marketplace app: request is not available for agreement")
	ErrQuoteUnavailable = errors.New("marketplace app: quote is not available for agreement")
	ErrRiderMismatch = errors.New("marketplace app: rider does not own request")
)

type AcceptQuoteCommand struct { QuoteID, RiderID, IdempotencyKey, AgreementID, EventID string; Now time.Time }
type OutboxEvent struct { ID, Name, AggregateType, AggregateID, InstanceID string; Payload map[string]any; OccurredAt time.Time }
type AcceptanceStore interface { WithinTx(context.Context, func(AcceptanceTx) error) error }
type AcceptanceTx interface {
	LockIdempotency(context.Context,string,string) error
	FindIdempotentAgreement(context.Context,string,string)(marketplace.Agreement,bool,error)
	GetQuoteForUpdate(context.Context,string)(marketplace.Quote,error)
	GetRequestForUpdate(context.Context,string)(marketplace.Request,error)
	InsertAgreement(context.Context,marketplace.Agreement) error
	MarkQuoteAccepted(context.Context,string,time.Time) error
	MarkRequestAgreed(context.Context,string) error
	InvalidateOtherQuotes(context.Context,string,string) error
	SaveIdempotentAgreement(context.Context,string,string,string) error
	AppendOutbox(context.Context,OutboxEvent) error
}
type AcceptanceService struct { Store AcceptanceStore }

func (s AcceptanceService) AcceptQuote(ctx context.Context, cmd AcceptQuoteCommand) (marketplace.Agreement,error) {
	if s.Store==nil||cmd.QuoteID==""||cmd.RiderID==""||cmd.IdempotencyKey==""||cmd.AgreementID==""||cmd.EventID==""||cmd.Now.IsZero(){return marketplace.Agreement{},ErrInvalidCommand}
	var result marketplace.Agreement
	err:=s.Store.WithinTx(ctx,func(tx AcceptanceTx) error {
		if err:=tx.LockIdempotency(ctx,cmd.RiderID,cmd.IdempotencyKey);err!=nil{return err}
		if existing,ok,err:=tx.FindIdempotentAgreement(ctx,cmd.RiderID,cmd.IdempotencyKey);err!=nil{return err}else if ok{result=existing;return nil}
		quote,err:=tx.GetQuoteForUpdate(ctx,cmd.QuoteID);if err!=nil{return err};if !quote.IsSelectable(cmd.Now){return ErrQuoteUnavailable}
		request,err:=tx.GetRequestForUpdate(ctx,quote.RequestID);if err!=nil{return err};if request.RiderID!=cmd.RiderID{return ErrRiderMismatch};if request.Status!=marketplace.RequestOpen&&request.Status!=marketplace.RequestReceivingQuotes{return ErrRequestUnavailable}
		agreement,err:=marketplace.NewAgreement(cmd.AgreementID,request,quote,cmd.Now);if err!=nil{return err}
		if err:=tx.InsertAgreement(ctx,agreement);err!=nil{return err}
		if err:=tx.MarkQuoteAccepted(ctx,quote.ID,cmd.Now);err!=nil{return err}
		if err:=tx.InvalidateOtherQuotes(ctx,request.ID,quote.ID);err!=nil{return err}
		if err:=tx.MarkRequestAgreed(ctx,request.ID);err!=nil{return err}
		if err:=tx.SaveIdempotentAgreement(ctx,cmd.RiderID,cmd.IdempotencyKey,agreement.ID);err!=nil{return err}
		if err:=tx.AppendOutbox(ctx,OutboxEvent{ID:cmd.EventID,Name:"marketplace.agreement.created.v1",AggregateType:"agreement",AggregateID:agreement.ID,InstanceID:agreement.InstanceID,Payload:map[string]any{"agreement_id":agreement.ID,"request_id":agreement.RequestID,"quote_id":agreement.QuoteID,"rider_id":agreement.RiderID,"driver_id":agreement.DriverID,"service_type":agreement.ServiceType,"currency":agreement.Fare.Currency,"fare_minor":agreement.Fare.Minor},OccurredAt:cmd.Now.UTC()});err!=nil{return err}
		result=agreement;return nil
	})
	if err!=nil{return marketplace.Agreement{},err};return result,nil
}
