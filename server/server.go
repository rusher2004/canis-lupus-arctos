package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rusher2004/canis-lupus-arctos/store"
)

type RiskStore interface {
	CreateRisk(state, title, desc string) (store.Risk, error)
	GetRisk(id string) (store.Risk, error)
	GetRiskList() ([]store.Risk, error)
}

type risk struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func errResponse(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	fmt.Fprintf(w, `{"error": "%s"}`, msg)
}

func jsonResponse(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		errResponse(w, "internal server error", http.StatusInternalServerError)
	}
}

// NewServer creates a new http.Handler to provide the REST API for the risk store.
func NewServer(rs RiskStore, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, rs, logger)

	return mux
}

func addRoutes(mux *http.ServeMux, rs RiskStore, logger *slog.Logger) {
	mux.Handle("POST /v1/risk",
		requestLogger(logger, dontPanic(logger, handleCreateRisk(rs))),
	)
	mux.Handle("GET /v1/risk/",
		requestLogger(logger, dontPanic(logger, handleGetRiskList(rs))),
	)
	mux.Handle("GET /v1/risk/{id}",
		requestLogger(logger, dontPanic(logger, handleGetRisk(rs))),
	)
}

func handleCreateRisk(rs RiskStore) http.HandlerFunc {
	type input struct {
		State       string `json:"state"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	states := []string{"open", "closed", "accepted", "investigating"}
	invalidMessage := fmt.Sprintf("state must be one of [%s]", strings.Join(states, ", "))

	return func(w http.ResponseWriter, r *http.Request) {
		var in input
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			if errors.Is(err, io.EOF) {
				errResponse(w, "missing body", http.StatusBadRequest)
				return
			}

			errResponse(w, "invalid request", http.StatusBadRequest)
			return
		}

		// only state is required
		if in.State == "" {
			errResponse(w, "state required", http.StatusBadRequest)
			return
		}

		if !slices.Contains(states, in.State) {
			errResponse(w, invalidMessage, http.StatusBadRequest)
			return
		}

		out, err := rs.CreateRisk(in.State, in.Title, in.Description)
		if err != nil {
			errResponse(w, "internal server error", http.StatusInternalServerError)
			return
		}

		jsonResponse(
			w,
			risk{ID: out.ID.String(), State: out.State, Title: out.Title, Description: out.Description},
			http.StatusCreated,
		)
	}
}

func handleGetRisk(rs RiskStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			errResponse(w, "invalid risk id", http.StatusBadRequest)
			return
		}

		out, err := rs.GetRisk(uid.String())
		if err != nil {
			errResponse(w, "risk not found", http.StatusNotFound)
			return
		}

		jsonResponse(
			w,
			risk{ID: out.ID.String(), State: out.State, Title: out.Title, Description: out.Description},
			http.StatusOK,
		)
	}
}

func handleGetRiskList(rs RiskStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		risks, err := rs.GetRiskList()
		if err != nil {
			errResponse(w, "internal server error", http.StatusInternalServerError)
			return
		}

		out := make([]risk, 0, len(risks))
		for _, r := range risks {
			out = append(out, risk{
				ID:          r.ID.String(),
				State:       r.State,
				Title:       r.Title,
				Description: r.Description,
			})
		}

		jsonResponse(w, out, http.StatusOK)
	}
}

// dontPanic is a middleware to ensure that the server does not crash due to a panic in the handler h.
// If a recover is triggered, it will log the error and return a 500 status code.
func dontPanic(logger *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logger.Error("recovered from panic", "error", rvr)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

// requestLogger is a middleware to log the request and response details.
func requestLogger(logger *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		logger.Info("request", "method", r.Method, "url", r.URL.String())
		defer func() {
			logger.Info("response", "url", r.URL.String(), "duration", time.Since(now).String())
		}()

		h.ServeHTTP(w, r)
	})
}
