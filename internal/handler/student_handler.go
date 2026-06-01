package handler

import (
	"net/http"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	studentService service.StudentService
}

func NewStudentHandler(studentService service.StudentService) *StudentHandler {
	return &StudentHandler{studentService: studentService}
}

// GetStudents godoc
// @Summary      Get all students (Admin)
// @Description  Retrieve all students from the database, ordered by name
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200      {array}   model.Student
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/students [get]
func (h *StudentHandler) GetStudents(c *gin.Context) {
	var onlyActive *bool
	if activeParam := c.Query("active"); activeParam != "" {
		active := activeParam == "true" || activeParam == "1"
		onlyActive = &active
	}

	res, err := h.studentService.GetAllStudents(c.Request.Context(), onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetStudent godoc
// @Summary      Get student by ID (Admin)
// @Description  Retrieve a specific student by their unique ID
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Student ID"
// @Success      200      {object}  model.Student
// @Failure      404      {object}  model.ErrorResponse
// @Router       /admin/students/{id} [get]
func (h *StudentHandler) GetStudent(c *gin.Context) {
	param := c.Param("id")
	var (
		res *model.Student
		err error
	)
	if id, ok := parseNumericID(param); ok {
		res, err = h.studentService.GetStudentByID(c.Request.Context(), id)
	} else {
		res, err = h.studentService.GetStudentByPublicID(c.Request.Context(), param)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Siswa tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// CreateStudent godoc
// @Summary      Create student (Admin)
// @Description  Add a new student record
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.StudentRequest  true  "Student Request"
// @Success      201      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/students [post]
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var req model.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	err := h.studentService.CreateStudent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.WebResponse{Message: "Siswa berhasil dibuat"})
}

// UpdateStudent godoc
// @Summary      Update student (Admin)
// @Description  Modify an existing student's information
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                   true  "Student ID"
// @Param        request  body      model.StudentRequest  true  "Student Request"
// @Success      200      {object}  model.WebResponse
// @Failure      400      {object}  model.ErrorResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/students/{id} [put]
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	var req model.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	param := c.Param("id")
	student, err := func() (*model.Student, error) {
		if id, ok := parseNumericID(param); ok {
			return h.studentService.GetStudentByID(c.Request.Context(), id)
		}
		return h.studentService.GetStudentByPublicID(c.Request.Context(), param)
	}()
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Siswa tidak ditemukan"})
		return
	}

	err = h.studentService.UpdateStudent(c.Request.Context(), student.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Siswa berhasil diupdate"})
}

// DeleteStudent godoc
// @Summary      Delete student (Admin)
// @Description  Remove a student and their associated sessions
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Student ID"
// @Success      200      {object}  model.WebResponse
// @Failure      500      {object}  model.ErrorResponse
// @Router       /admin/students/{id} [delete]
func (h *StudentHandler) DeleteStudent(c *gin.Context) {
	param := c.Param("id")
	student, err := func() (*model.Student, error) {
		if id, ok := parseNumericID(param); ok {
			return h.studentService.GetStudentByID(c.Request.Context(), id)
		}
		return h.studentService.GetStudentByPublicID(c.Request.Context(), param)
	}()
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Siswa tidak ditemukan"})
		return
	}

	err = h.studentService.DeleteStudent(c.Request.Context(), student.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WebResponse{Message: "Siswa berhasil dihapus"})
}
