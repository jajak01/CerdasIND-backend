package handler

import (
	"net/http"
	"strconv"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	documentService service.DocumentService
}

func NewDocumentHandler(documentService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

func parseDocumentFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})
	if studentID := c.Query("studentId"); studentID != "" {
		if id, err := strconv.ParseInt(studentID, 10, 64); err == nil {
			filters["student_id"] = id
		}
	}
	if linkedInvoiceID := c.Query("linkedInvoiceId"); linkedInvoiceID != "" {
		if id, err := strconv.ParseInt(linkedInvoiceID, 10, 64); err == nil {
			filters["linked_invoice_id"] = id
		}
	}
	return filters
}

func (h *DocumentHandler) CreateInvoice(c *gin.Context) {
	var req model.StudentDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int64)
	res, err := h.documentService.CreateInvoice(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.WebResponse{Message: "Invoice berhasil disimpan", Data: res})
}

func (h *DocumentHandler) CreateReport(c *gin.Context) {
	var req model.StudentDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int64)
	res, err := h.documentService.CreateReport(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.WebResponse{Message: "Report berhasil disimpan", Data: res})
}

func (h *DocumentHandler) GetInvoices(c *gin.Context) {
	res, err := h.documentService.GetInvoices(c.Request.Context(), parseDocumentFilters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *DocumentHandler) GetReports(c *gin.Context) {
	res, err := h.documentService.GetReports(c.Request.Context(), parseDocumentFilters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *DocumentHandler) GetInvoice(c *gin.Context) {
	param := c.Param("id")
	var (
		res *model.StudentDocument
		err error
	)
	if id, ok := parseNumericID(param); ok {
		res, err = h.documentService.GetInvoiceByID(c.Request.Context(), id)
	}
	if err != nil || res == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Invoice tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *DocumentHandler) GetReport(c *gin.Context) {
	param := c.Param("id")
	var (
		res *model.StudentDocument
		err error
	)
	if id, ok := parseNumericID(param); ok {
		res, err = h.documentService.GetReportByID(c.Request.Context(), id)
	}
	if err != nil || res == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Report tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, res)
}
