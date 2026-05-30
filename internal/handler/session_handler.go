package handler

import (
	"net/http"
	"strconv"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService service.SessionService
}

func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

// GetSessions godoc
// @Summary      Get all sessions (Admin)
// @Description  Retrieve sessions with support for various filters
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        studentId      query     int     false  "Student ID"
// @Param        startDate      query     string  false  "Start Date (YYYY-MM-DD)"
// @Param        endDate        query     string  false  "End Date (YYYY-MM-DD)"
// @Param        status         query     string  false  "Session Status (scheduled, completed, cancelled)"
// @Param        paymentStatus  query     string  false  "Payment Status (pending, paid, overdue)"
// @Param        search         query     string  false  "Search by student name or contact"
// @Success      200      {array}   model.Session
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/sessions [get]
func (h *SessionHandler) GetSessions(c *gin.Context) {
	filters := make(map[string]interface{})
	if studentID := c.Query("studentId"); studentID != "" {
		id, _ := strconv.ParseInt(studentID, 10, 64)
		filters["student_id"] = id
	}
	filters["start_date"] = c.Query("startDate")
	filters["end_date"] = c.Query("endDate")
	filters["status"] = c.Query("status")
	filters["payment_status"] = c.Query("paymentStatus")
	filters["search"] = c.Query("search")

	res, err := h.sessionService.GetAllSessions(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetSession godoc
// @Summary      Get session by ID (Admin)
// @Description  Retrieve a specific session including the associated student details
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Session ID"
// @Success      200      {object}  model.Session
// @Failure      404      {object}  model.ErrorResponse
// @Router       /admin/sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.sessionService.GetSessionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Sesi tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// CreateSession godoc
// @Summary      Create session (Admin)
// @Description  Record a new tutoring session
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.SessionRequest  true  "Session Request"
// @Success      201      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req model.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	err := h.sessionService.CreateSession(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.WebResponse{Message: "Sesi berhasil dibuat"})
}

// UpdateSession godoc
// @Summary      Update session (Admin)
// @Description  Update session details. Handles logic for payment_date
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                   true  "Session ID"
// @Param        request  body      model.SessionRequest  true  "Session Request"
// @Success      200      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/sessions/{id} [put]
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	err := h.sessionService.UpdateSession(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Sesi berhasil diupdate"})
}

// DeleteSession godoc
// @Summary      Delete session (Admin)
// @Description  Remove a session record
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Session ID"
// @Success      200      {object}  model.WebResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.sessionService.DeleteSession(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Sesi berhasil dihapus"})
}

// GetDashboardStats godoc
// @Summary      Get dashboard stats (Admin)
// @Description  Calculates key performance indicators (KPIs) for the dashboard
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  model.DashboardStats
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/dashboard/stats [get]
func (h *SessionHandler) GetDashboardStats(c *gin.Context) {
	res, err := h.sessionService.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
