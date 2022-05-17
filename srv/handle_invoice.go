package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HandleInvoice creates invoices
func (s *Server) HandleInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBytesRead)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createInvoiceReq
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		// Request body must only contain a single JSON object
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := s.createInvoice(ctx, &req); err != nil {
		if errors.Is(err, ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

type createInvoiceReq struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
	Label  string  `json:"label"`
}

func (s *Server) validateCreateInvoiceReq(ctx context.Context, req *createInvoiceReq) error {
	if req.UserID <= 0 {
		return fmt.Errorf("%w: bad initial user ID: %d", ErrBadRequest, req.UserID)
	}
	if req.Amount <= 99.99 {
		return fmt.Errorf("%w: bad amount: %v", ErrBadRequest, req.Amount)
	}
	if req.Label == "" {
		return fmt.Errorf("%w: empty label", ErrBadRequest)
	}
	if _, err := s.userForID(ctx, req.UserID); err != nil {
		return fmt.Errorf("%w: user %d does not exist", ErrBadRequest, req.UserID)
	}
	return nil
}

func (s *Server) createInvoice(ctx context.Context, req *createInvoiceReq) (err error) {
	log := NewLogFromCtx(ctx)
	log.Info("handling createInvoice")
	start := time.Now()

	if err = s.validateCreateInvoiceReq(ctx, req); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	pg := s.newPGClient(ctx)

	q := pg.Insert("invoices").
		// NOTE: possible race condition WRT existence of user_id
		SetMap(map[string]interface{}{
			"user_id": req.UserID,
			"amount":  int64(req.Amount * 100.0),
			"label":   req.Label,
		}).
		Suffix("RETURNING id, status")

	var rows *sql.Rows
	if rows, err = q.QueryContext(ctx); err != nil {
		log.Error("", zap.Error(err))
		return
	}
	defer rows.Close()
	var invoiceID int64
	var invoiceStatus string
	for rows.Next() {
		if err = rows.Scan(&invoiceID, &invoiceStatus); err != nil {
			log.Error("", zap.Error(err))
			return
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	if invoiceID <= 0 || invoiceStatus != "pending" {
		err = fmt.Errorf("%w: unknown user_id %v", ErrBadRequest, req.UserID)
		log.Error("", zap.Error(err))
		return
	}

	log.Info("handled createInvoice",
		zap.Int64("invoiceID", invoiceID),
		zap.String("invoiceStatus", invoiceStatus),
		zap.Duration("in", time.Since(start)),
	)
	return
}
