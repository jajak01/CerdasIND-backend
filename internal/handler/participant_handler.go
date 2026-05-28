package handler

import (
	"log"
	"net/http"
	"strconv"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ParticipantHandler struct {
	participantService service.ParticipantService
}

func NewParticipantHandler(participantService service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{participantService: participantService}
}

// GetJenjang godoc
// @Summary      Get all jenjang
// @Description  Retrieve list of all educational levels (SD, SMP, SMA)
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Success      200      {array}   model.Jenjang
// @Failure      500      {object}  model.ErrorResponse
// @Router       /jenjang [get]
func (h *ParticipantHandler) GetJenjang(c *gin.Context) {
	res, err := h.participantService.GetJenjang(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetMapel godoc
// @Summary      Get mapel by jenjang
// @Description  Retrieve list of subjects based on jenjang ID
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Jenjang ID"
// @Success      200      {array}   model.Mapel
// @Failure      500      {object}  model.ErrorResponse
// @Router       /jenjang/{id}/mapel [get]
func (h *ParticipantHandler) GetMapel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	res, err := h.participantService.GetMapelByJenjang(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetBundles godoc
// @Summary      Get bundles by mapel
// @Description  Retrieve list of active exam bundles based on mapel ID
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Mapel ID"
// @Success      200      {array}   model.Bundle
// @Failure      500      {object}  model.ErrorResponse
// @Router       /mapel/{id}/bundles [get]
func (h *ParticipantHandler) GetBundles(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	res, err := h.participantService.GetBundlesByMapel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetSoal godoc
// @Summary      Get soal by bundle
// @Description  Retrieve questions for an exam bundle (hides keys and discussion)
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bundle ID"
// @Success      200      {array}   model.SoalPublic
// @Failure      500      {object}  model.ErrorResponse
// @Router       /bundles/{id}/soal [get]
func (h *ParticipantHandler) GetSoal(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.MustGet("user_id").(int64)
	res, err := h.participantService.GetSoalByBundle(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Submit godoc
// @Summary      Submit exam answers
// @Description  Submit answers for an exam bundle and get auto-grade result
// @Tags         participant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                 true  "Bundle ID"
// @Param        request  body      model.SubmitRequest  true  "Submit Request"
// @Success      200      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /bundles/{id}/submit [post]
func (h *ParticipantHandler) Submit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.MustGet("user_id").(int64)

	var req model.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	res, err := h.participantService.SubmitUjian(c.Request.Context(), userID, id, req)
	if err != nil {
		if err.Error() == "ujian sudah disubmit" {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetHistory godoc
// @Summary      Get user history
// @Description  Retrieve exam history for the logged-in user
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Success      200      {array}   model.HistoryResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /users/history [get]
func (h *ParticipantHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	res, err := h.participantService.GetHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetReview godoc
// @Summary      Get exam review
// @Description  Retrieve questions with keys and discussion after exam is finished
// @Tags         participant
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bundle ID"
// @Success      200      {array}   model.ReviewResponse
// @Failure      403      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /bundles/{id}/review [get]
func (h *ParticipantHandler) GetReview(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.MustGet("user_id").(int64)
	res, err := h.participantService.GetReview(c.Request.Context(), userID, id)
	if err != nil {
		log.Printf("Error in GetReview for user %d bundle %d: %v", userID, id, err)
		if err.Error() == "akses ditolak: status masih menunggu koreksi" {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Message: err.Error()})
		} else if err.Error() == "not found: history ujian tidak ditemukan" {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "History ujian tidak ditemukan"})
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}
