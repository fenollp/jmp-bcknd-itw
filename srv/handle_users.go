package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HandleUsers lists users
func (s *Server) HandleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBytesRead)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	req := &listUsersReq{count: 50}
	rep, err := s.listUsers(ctx, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(rep); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type listUsersReq struct {
	fromID int64
	count  int
}

func (s *Server) validateListUsersReq(ctx context.Context, req *listUsersReq) error {
	if req.fromID < 0 {
		return fmt.Errorf("bad initial user ID: %v", req.fromID)
	}
	if req.count <= 0 {
		return fmt.Errorf("bad count: %v", req.count)
	}
	return nil
}

func (s *Server) listUsers(ctx context.Context, req *listUsersReq) (users []User, err error) {
	log := NewLogFromCtx(ctx)
	log.Info("handling listUsers")
	start := time.Now()

	if err = s.validateListUsersReq(ctx, req); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	pg := s.newPGClient(ctx)

	// Let's assume users can be added / modified through an API unrelated
	// to this service. Meaning we have no simple way to cache these calls.
	if users, err = pg.readUsers(ctx, req); err != nil {
		log.Error("", zap.Error(err))
		return
	}

	log.Info("handled listUsers",
		zap.Int("#", len(users)),
		zap.Duration("in", time.Since(start)),
	)
	return
}
