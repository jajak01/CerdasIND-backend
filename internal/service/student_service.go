package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"cerdasind-backend/pkg/utils"
	"github.com/jmoiron/sqlx"
)

type StudentService interface {
	CreateStudent(ctx context.Context, req model.StudentRequest) error
	UpdateStudent(ctx context.Context, id int64, req model.StudentRequest) error
	DeleteStudent(ctx context.Context, id int64) error
	GetStudentByID(ctx context.Context, id int64) (*model.Student, error)
	GetStudentByPublicID(ctx context.Context, publicID string) (*model.Student, error)
	GetAllStudents(ctx context.Context, onlyActive *bool) ([]model.Student, error)
}

type studentService struct {
	db          *sqlx.DB
	studentRepo repository.StudentRepository
	userRepo    repository.UserRepository
}

func NewStudentService(db *sqlx.DB, studentRepo repository.StudentRepository, userRepo repository.UserRepository) StudentService {
	return &studentService{
		db:          db,
		studentRepo: studentRepo,
		userRepo:    userRepo,
	}
}

func (s *studentService) CreateStudent(ctx context.Context, req model.StudentRequest) error {
	// Start Transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := repository.InjectTx(ctx, tx)

	var userID *int64

	// Auto-generate User if user_id is not provided (or 0)
	if req.UserID == 0 {
		// Generate simple username: name without spaces + random suffix or just name
		username := strings.ToLower(strings.ReplaceAll(req.Name, " ", ""))
		email := username + "@cerdasind.com" // Default email

		// Check if user already exists with this email (simple handle)
		existing, _ := s.userRepo.FindByEmail(txCtx, email)
		if existing != nil {
			email = fmt.Sprintf("%s%d@cerdasind.com", username, time.Now().Unix())
		}

		// Default password (should be changed later by student)
		hash, _ := utils.HashPassword("password123")

		user := &model.User{
			Username:     username,
			Email:        email,
			PasswordHash: hash,
			Role:         model.RolePeserta,
		}

		err = s.userRepo.Create(txCtx, user)
		if err != nil {
			return fmt.Errorf("gagal generate user: %v", err)
		}
		userID = &user.ID
	} else {
		userID = &req.UserID
	}

	student := &model.Student{
		UserID:   userID,
		Name:     req.Name,
		School:   req.School,
		Grade:    req.Grade,
		Contact:  req.Contact,
		Address:  req.Address,
		IsActive: true,
	}
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	err = s.studentRepo.Create(txCtx, student)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *studentService) UpdateStudent(ctx context.Context, id int64, req model.StudentRequest) error {
	student, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	student.Name = req.Name
	student.School = req.School
	student.Grade = req.Grade
	student.Contact = req.Contact
	student.Address = req.Address
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	return s.studentRepo.Update(ctx, student)
}

func (s *studentService) DeleteStudent(ctx context.Context, id int64) error {
	return s.studentRepo.Delete(ctx, id)
}

func (s *studentService) GetStudentByID(ctx context.Context, id int64) (*model.Student, error) {
	return s.studentRepo.FindByID(ctx, id)
}

func (s *studentService) GetStudentByPublicID(ctx context.Context, publicID string) (*model.Student, error) {
	return s.studentRepo.FindByPublicID(ctx, publicID)
}

func (s *studentService) GetAllStudents(ctx context.Context, onlyActive *bool) ([]model.Student, error) {
	return s.studentRepo.FindAll(ctx, onlyActive)
}
