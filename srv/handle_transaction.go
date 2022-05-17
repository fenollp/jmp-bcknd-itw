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

// HandleTransaction completes an invoice
func (s *Server) HandleTransaction(w http.ResponseWriter, r *http.Request) {
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

	var req createTransactionReq
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
	if err := s.createTransaction(ctx, &req); err != nil {
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

type createTransactionReq struct {
	InvoiceID int64   `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	Reference string  `json:"reference"`
}

func (s *Server) validateCreateTransactionReq(ctx context.Context, req *createTransactionReq) error {
	if req.InvoiceID <= 0 {
		return fmt.Errorf("%w: bad invoice ID: %d", ErrBadRequest, req.InvoiceID)
	}
	if req.Amount <= 99.99 {
		return fmt.Errorf("%w: bad amount: %v", ErrBadRequest, req.Amount)
	}
	if req.Reference == "" {
		return fmt.Errorf("%w: empty reference", ErrBadRequest)
	}
	return nil
}

func (s *Server) createTransaction(ctx context.Context, req *createTransactionReq) (err error) {
	log := NewLogFromCtx(ctx)
	log.Info("handling createTransaction")
	start := time.Now()

	if err = s.validateCreateTransactionReq(ctx, req); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	var tx *sql.Tx
	if tx, err = s.pg.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	// Pay the user of the matching pending invoice

	var result sql.Result
	if result, err = tx.ExecContext(ctx, `
UPDATE users
SET balance = balance + $1
WHERE id = ( SELECT users.id
			 FROM users
			 JOIN invoices ON users.id = invoices.user_id
			 WHERE invoices.id = $2
			 AND invoices.amount = $1
			 AND invoices.status = 'pending'
			)`,
		int64(req.Amount*100.0),
		req.InvoiceID,
	); err != nil {
		log.Error("", zap.Error(err))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn("", zap.Error(err))
		}
		return
	}
	var rowsAffected int64
	if rowsAffected, err = result.RowsAffected(); err != nil {
		log.Error("", zap.Error(err))
		return
	}
	if rowsAffected != 1 {
		err = fmt.Errorf("%w: unusable invoice %+v", ErrBadRequest, req)
		log.Error("", zap.Error(err))
		return
	}

	// Set the matching pending invoice status to paid

	var rows *sql.Rows
	if rows, err = tx.QueryContext(ctx, `
UPDATE invoices
SET status = 'paid'
WHERE id = $1
AND amount = $2
AND status = 'pending'
RETURNING id
`,
		req.InvoiceID,
		int64(req.Amount*100.0),
	); err != nil {
		log.Error("", zap.Error(err))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn("", zap.Error(err))
		}
		return
	}
	defer rows.Close()
	var invoiceID int64
	for rows.Next() {
		if err = rows.Scan(&invoiceID); err != nil {
			log.Error("", zap.Error(err))
			return
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("", zap.Error(err))
		return
	}
	if invoiceID <= 0 || invoiceID != req.InvoiceID {
		err = fmt.Errorf("%w: unknown invoice %+v", ErrBadRequest, req)
		log.Error("", zap.Error(err))
		return
	}

	if err = tx.Commit(); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	log.Info("handled createTransaction",
		zap.Int64("invoiceID", invoiceID),
		zap.Duration("in", time.Since(start)),
	)
	return
}
