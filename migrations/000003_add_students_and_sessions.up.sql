-- Create Enums for Session Status and Payment Status
CREATE TYPE session_status AS ENUM ('scheduled', 'completed', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'overdue');

-- Tabel Students (Siswa)
-- Relation: Students are users, but not all users are students.
-- user_id is nullable to allow student creation without immediate user account, 
-- but linked to users table for future authentication and profile management.
CREATE TABLE students (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    school VARCHAR(255),
    grade VARCHAR(50),
    contact VARCHAR(50),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel Sessions (Sesi Les)
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    time TIME NOT NULL,
    subject VARCHAR(100) DEFAULT 'isikan mapel' NOT NULL,
    notes TEXT,
    price NUMERIC(12,2) DEFAULT 20000.00 NOT NULL,
    status session_status DEFAULT 'scheduled' NOT NULL,
    payment_status payment_status DEFAULT 'pending' NOT NULL,
    payment_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for performance
CREATE INDEX idx_sessions_student ON sessions(student_id);
CREATE INDEX idx_sessions_date ON sessions(date);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_payment_status ON sessions(payment_status);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_students_updated_at BEFORE UPDATE ON students FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
