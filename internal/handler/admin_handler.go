package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GetBundles godoc
// @Summary      Get all bundles (Admin)
// @Description  Retrieve all exam bundles including inactive ones
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200      {array}   model.Bundle
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/bundles [get]
func (h *AdminHandler) GetBundles(c *gin.Context) {
	res, err := h.adminService.GetAllBundles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UploadBundle godoc
// @Summary      Upload bundle and soal from Excel
// @Description  Create a new bundle and bulk insert questions from .xlsx file
// @Tags         admin
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file         formData  file    true  "Excel File"
// @Param        mapel_id     formData  int     true  "Mapel ID"
// @Param        nama_bundle  formData  string  true  "Bundle Name"
// @Param        waktu_menit  formData  int     true  "Time in Minutes"
// @Success      201      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/bundles/upload [post]
func (h *AdminHandler) UploadBundle(c *gin.Context) {
	file, _ := c.FormFile("file")
	mapelID, _ := strconv.Atoi(c.PostForm("mapel_id"))
	namaBundle := c.PostForm("nama_bundle")
	waktuMenit, _ := strconv.Atoi(c.PostForm("waktu_menit"))
	userID := c.MustGet("user_id").(int64)

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}
	defer f.Close()

	err = h.adminService.UploadBundle(c.Request.Context(), f, mapelID, namaBundle, waktuMenit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.WebResponse{Message: "Bundle berhasil dibuat dan soal berhasil di-insert."})
}

// ExportBundle godoc
// @Summary      Export soal to Excel
// @Description  Download all questions in a bundle as .xlsx file
// @Tags         admin
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        id   path      int  true  "Bundle ID"
// @Success      200      {file}    binary
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/bundles/{id}/export [get]
func (h *AdminHandler) ExportBundle(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.adminService.ExportBundle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=bundle_%d.xlsx", id))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", res)
}

// UpdateBundle godoc
// @Summary      Update soal via Excel
// @Description  Update existing questions or add new ones via .xlsx upload
// @Tags         admin
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int   true  "Bundle ID"
// @Param        file  formData  file  true  "Excel File"
// @Success      200      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/bundles/{id}/update [put]
func (h *AdminHandler) UpdateBundle(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	file, _ := c.FormFile("file")

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}
	defer f.Close()

	err = h.adminService.UpdateBundleWithExcel(c.Request.Context(), id, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Update berhasil."})
}

// GetSubmissions godoc
// @Summary      Get user submissions
// @Description  Retrieve list of student exam submissions, filtered by status
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        status  query     string  false  "Status (menunggu_koreksi, selesai)"
// @Success      200      {array}   map[string]interface{}
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/submissions [get]
func (h *AdminHandler) GetSubmissions(c *gin.Context) {
	status := c.Query("status")
	res, err := h.adminService.GetSubmissions(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetSubmissionDetail godoc
// @Summary      Get submission detail
// @Description  Retrieve detailed answers of a student submission for grading
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        history_id  path      int  true  "History ID"
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/submissions/{history_id} [get]
func (h *AdminHandler) GetSubmissionDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("history_id"), 10, 64)
	res, err := h.adminService.GetSubmissionDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GradeSubmission godoc
// @Summary      Grade essay questions
// @Description  Submit manual scores for essay questions and finalize exam
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        history_id  path      int                 true  "History ID"
// @Param        request     body      model.GradeRequest  true  "Grade Request"
// @Success      200      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/submissions/{history_id}/grade [put]
func (h *AdminHandler) GradeSubmission(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("history_id"), 10, 64)
	var req model.GradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	err := h.adminService.GradeSubmission(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Penilaian berhasil. Status ujian selesai."})
}
