package handler

import (
	"errors"
	"net/http"
	"schedule/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc service.Svc
}

func New(svc service.Svc) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetUnscheduledQuotes(c *gin.Context) {
	quotes, err := h.svc.GetUnscheduledQuotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

func (h *Handler) CreateJob(c *gin.Context) {
	var input service.CreateJobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.svc.CreateJob(input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStartInPast):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrConflict), errors.Is(err, service.ErrQuoteAlreadyAssigned):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (h *Handler) CompleteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}

	if err := h.svc.CompleteJob(uint(id)); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrAlreadyDone):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job completed"})
}

func (h *Handler) GetTechnicianNotifications(c *gin.Context) {
	h.getNotifications(c, "technician")
}

func (h *Handler) GetManagerNotifications(c *gin.Context) {
	h.getNotifications(c, "manager")
}

func (h *Handler) getNotifications(c *gin.Context, userType string) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	notifications, err := h.svc.GetNotifications(userType, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}
