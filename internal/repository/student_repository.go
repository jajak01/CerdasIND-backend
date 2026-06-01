package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"cerdasind-backend/pkg/utils"
	"github.com/jmoiron/sqlx"
)

type StudentRepository interface {
	Create(ctx context.Context, student *model.Student) error
	Update(ctx context.Context, student *model.Student) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*model.Student, error)
	FindByPublicID(ctx context.Context, publicID string) (*model.Student, error)
	FindByUserID(ctx context.Context, userID int64) (*model.Student, error)
	FindAll(ctx context.Context, onlyActive *bool) ([]model.Student, error)
	Count(ctx context.Context) (int, error)
}

type studentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) Create(ctx context.Context, student *model.Student) error {
	if student.PublicID == "" {
		publicID, err := utils.GenerateUUID()
		if err != nil {
			return err
		}
		student.PublicID = publicID
	}

	query := `INSERT INTO students (public_id, user_id, name, school, grade, contact, address, is_active) 
              VALUES ($1, NULLIF($2, 0), $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, 
		student.PublicID, student.UserID, student.Name, student.School, student.Grade, student.Contact, student.Address, student.IsActive).
		Scan(&student.ID, &student.CreatedAt, &student.UpdatedAt)
}

func (r *studentRepository) Update(ctx context.Context, student *model.Student) error {
	query := `UPDATE students SET name = $1, school = $2, grade = $3, contact = $4, address = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP 
              WHERE id = $7 RETURNING updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, 
		student.Name, student.School, student.Grade, student.Contact, student.Address, student.IsActive, student.ID).
		Scan(&student.UpdatedAt)
}

func (r *studentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := getRunner(ctx, r.db).ExecContext(ctx, query, id)
	return err
}

func (r *studentRepository) FindByID(ctx context.Context, id int64) (*model.Student, error) {
	var student model.Student
	query := `SELECT * FROM students WHERE id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &student, query, id)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindByPublicID(ctx context.Context, publicID string) (*model.Student, error) {
	var student model.Student
	query := `SELECT * FROM students WHERE public_id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &student, query, publicID)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindByUserID(ctx context.Context, userID int64) (*model.Student, error) {
	var student model.Student
	query := `SELECT * FROM students WHERE user_id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &student, query, userID)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindAll(ctx context.Context, onlyActive *bool) ([]model.Student, error) {
	var students []model.Student
	query := `SELECT * FROM students`
	args := []interface{}{}
	if onlyActive != nil {
		query += ` WHERE is_active = $1`
		args = append(args, *onlyActive)
	}
	query += ` ORDER BY name ASC`
	err := getRunner(ctx, r.db).SelectContext(ctx, &students, query, args...)
	return students, err
}

func (r *studentRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM students`
	err := getRunner(ctx, r.db).GetContext(ctx, &count, query)
	return count, err
}
