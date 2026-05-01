package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"schedule/internal/handler"
	"schedule/internal/model"
	"schedule/internal/service"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type mockSvc struct {
	getUnscheduledQuotes func() ([]model.Quote, error)
	createJob            func(input service.CreateJobInput) (*model.Job, error)
	completeJob          func(id uint) error
	getNotifications     func(userType string, userID uint) ([]model.Notification, error)
}

func (m *mockSvc) GetUnscheduledQuotes() ([]model.Quote, error) { return m.getUnscheduledQuotes() }
func (m *mockSvc) CreateJob(input service.CreateJobInput) (*model.Job, error) {
	return m.createJob(input)
}
func (m *mockSvc) CompleteJob(id uint) error { return m.completeJob(id) }
func (m *mockSvc) GetNotifications(userType string, userID uint) ([]model.Notification, error) {
	return m.getNotifications(userType, userID)
}

func setupRouter(svc service.Svc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.New(svc)
	r := gin.New()
	r.GET("/quotes", h.GetUnscheduledQuotes)
	r.POST("/jobs", h.CreateJob)
	r.PATCH("/jobs/:id/complete", h.CompleteJob)
	r.GET("/technicians/:id/notifications", h.GetTechnicianNotifications)
	r.GET("/managers/:id/notifications", h.GetManagerNotifications)
	return r
}

func TestGetUnscheduledQuotes_OK(t *testing.T) {
	svc := &mockSvc{
		getUnscheduledQuotes: func() ([]model.Quote, error) {
			return []model.Quote{{ID: 1, Title: "Quote A", Status: "unscheduled"}}, nil
		},
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/quotes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var quotes []model.Quote
	if err := json.Unmarshal(w.Body.Bytes(), &quotes); err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].Title != "Quote A" {
		t.Fatalf("unexpected quotes: %+v", quotes)
	}
}

func TestCreateJob_BadRequest_MissingField(t *testing.T) {
	r := setupRouter(&mockSvc{})
	body := `{"quote_id":1}` // technician_id / manager_id / start_at が欠けている
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateJob_Conflict(t *testing.T) {
	svc := &mockSvc{
		createJob: func(input service.CreateJobInput) (*model.Job, error) {
			return nil, service.ErrConflict
		},
	}
	r := setupRouter(svc)
	body := `{"quote_id":1,"technician_id":1,"manager_id":1,"start_at":"2026-05-01T09:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateJob_QuoteAlreadyAssigned(t *testing.T) {
	svc := &mockSvc{
		createJob: func(input service.CreateJobInput) (*model.Job, error) {
			return nil, service.ErrQuoteAlreadyAssigned
		},
	}
	r := setupRouter(svc)
	body := `{"quote_id":1,"technician_id":1,"manager_id":1,"start_at":"2026-05-01T09:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateJob_Success(t *testing.T) {
	start := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	svc := &mockSvc{
		createJob: func(input service.CreateJobInput) (*model.Job, error) {
			return &model.Job{
				ID: 1, QuoteID: 1, TechnicianID: 1, ManagerID: 1,
				StartAt: start, EndAt: start.Add(2 * time.Hour), Status: "scheduled",
			}, nil
		},
	}
	r := setupRouter(svc)
	body := `{"quote_id":1,"technician_id":1,"manager_id":1,"start_at":"2026-05-01T09:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCompleteJob_NotFound(t *testing.T) {
	svc := &mockSvc{
		completeJob: func(id uint) error { return service.ErrNotFound },
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodPatch, "/jobs/99/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCompleteJob_AlreadyDone(t *testing.T) {
	svc := &mockSvc{
		completeJob: func(id uint) error { return service.ErrAlreadyDone },
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodPatch, "/jobs/1/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCompleteJob_Success(t *testing.T) {
	svc := &mockSvc{
		completeJob: func(id uint) error { return nil },
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodPatch, "/jobs/1/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetTechnicianNotifications_InvalidID(t *testing.T) {
	r := setupRouter(&mockSvc{})
	req := httptest.NewRequest(http.MethodGet, "/technicians/abc/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetTechnicianNotifications_OK(t *testing.T) {
	svc := &mockSvc{
		getNotifications: func(userType string, userID uint) ([]model.Notification, error) {
			return []model.Notification{
				{ID: 1, UserType: "technician", UserID: 1, Message: "新しいJobが割り当てられました", JobID: 1},
			}, nil
		},
	}
	r := setupRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/technicians/1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
