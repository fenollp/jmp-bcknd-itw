package srv

import (
	"encoding/json"
	"io"
	"net/http"
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
	var req struct {
		// "user_id": 12,
		// "first_name": "Kevin",
		// "last_name": "Findus",
		// "balance": 492.97
		//FIXME
	}
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		// Request body must only contain a single JSON object
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ctx := r.Context()
	// //FIXME
	// // rep, err := s.ListUserCampaigns(ctx, &req)
	// // switch status.Code(err) {
	// // case codes.OK:
	// // case codes.PermissionDenied:
	// // 	w.WriteHeader(http.StatusForbidden)
	// // 	return
	// // default:
	// // 	w.WriteHeader(http.StatusBadRequest)
	// // 	return
	// // }

	// if err := json.NewEncoder(w).Encode(rep); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
}
