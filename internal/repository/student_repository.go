package repository

import (
	"context"
	"cerdasind-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type StudentRepository interface {
	Create(ctx context.Context, student *model.Student) error
	Update(ctx context.Context, student *model.Student) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*model.Student, error)
	FindByUserID(ctx context.Context, userID int64) (*model.Student, error)
	FindAll(ctx context.Context) ([]model.Student, error)
	Count(ctx context.Context) (int, error)
}

type studentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) Create(ctx context.Context, student *model.Student) error {
	query := `INSERT INTO students (user_id, name, school, grade, contact, address) 
              VALUES (NULLIF($1, 0), $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, 
		student.UserID, student.Name, student.School, student.Grade, student.Contact, student.Address).
		Scan(&student.ID, &student.CreatedAt, &student.UpdatedAt)
}

func (r *studentRepository) Update(ctx context.Context, student *model.Student) error {
	query := `UPDATE students SET name = $1, school = $2, grade = $3, contact = $4, address = $5, updated_at = CURRENT_TIMESTAMP 
              WHERE id = $6 RETURNING updated_at`
	return getRunner(ctx, r.db).QueryRowContext(ctx, query, 
		student.Name, student.School, student.Grade, student.Contact, student.Address, student.ID).
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

func (r *studentRepository) FindByUserID(ctx context.Context, userID int64) (*model.Student, error) {
	var student model.Student
	query := `SELECT * FROM students WHERE user_id = $1`
	err := getRunner(ctx, r.db).GetContext(ctx, &student, query, userID)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindAll(ctx context.Context) ([]model.Student, error) {
	var students []model.Student
	query := `SELECT * FROM students ORDER BY name ASC`
	err := getRunner(ctx, r.db).SelectContext(ctx, &students, query)
	return students, err
}

func (r *studentRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM students`
	err := getRunner(ctx, r.db).GetContext(ctx, &count, query)
	return count, err
}
