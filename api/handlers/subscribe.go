package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/immortalvibes/api/email"
)

const discountCode = "VIBE10"

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SubscribeStore is the DB subset needed by SubscribeHandler.
type SubscribeStore interface {
	SaveSubscriber(ctx context.Context, emailAddr string) (isNew bool, err error)
}

// SubscribeEmailer is the email subset needed by SubscribeHandler.
type SubscribeEmailer interface {
	SendDiscountCode(ctx context.Context, toEmail, code string) error
}

// SubscribeHandler handles POST /api/subscribe.
type SubscribeHandler struct {
	db      SubscribeStore
	emailer SubscribeEmailer
}

func NewSubscribeHandler(db SubscribeStore, emailer SubscribeEmailer) *SubscribeHandler {
	return &SubscribeHandler{db: db, emailer: emailer}
}

type subscribeRequest struct {
	Email string `json:"email"`
}

type subscribeResponse struct {
	Code string `json:"code"`
}

func (h *SubscribeHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !emailRe.MatchString(req.Email) {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}

	if _, err := h.db.SaveSubscriber(r.Context(), req.Email); err != nil {
		http.Error(w, "failed to save subscriber", http.StatusInternalServerError)
		return
	}

	// Always send the code — whether new or returning subscriber.
	if err := h.emailer.SendDiscountCode(r.Context(), req.Email, discountCode); err != nil {
		// Non-fatal: code is still returned in the response.
		_ = err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscribeResponse{Code: discountCode})
}

// Ensure email.Sender satisfies SubscribeEmailer at compile time.
var _ SubscribeEmailer = (*email.Sender)(nil)
