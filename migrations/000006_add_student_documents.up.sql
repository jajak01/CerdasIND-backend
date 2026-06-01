CREATE TYPE student_document_kind AS ENUM ('invoice', 'report');

CREATE TABLE student_documents (
    id BIGSERIAL PRIMARY KEY,
    public_id VARCHAR(36) UNIQUE NOT NULL,
    document_kind student_document_kind NOT NULL,
    document_number VARCHAR(50) UNIQUE NOT NULL,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    linked_invoice_id BIGINT REFERENCES student_documents(id) ON DELETE SET NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_amount NUMERIC(12,2) DEFAULT 0 NOT NULL,
    summary TEXT,
    message TEXT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE student_document_sessions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES student_documents(id) ON DELETE CASCADE,
    session_id BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    session_date DATE NOT NULL,
    session_time TIME NOT NULL,
    subject VARCHAR(100) NOT NULL,
    note TEXT,
    price NUMERIC(12,2) DEFAULT 0 NOT NULL,
    payment_status payment_status NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_student_documents_number ON student_documents(document_number);
CREATE INDEX idx_student_documents_kind ON student_documents(document_kind);
CREATE INDEX idx_student_documents_student ON student_documents(student_id);
CREATE INDEX idx_student_documents_linked_invoice ON student_documents(linked_invoice_id);
CREATE INDEX idx_student_document_sessions_document ON student_document_sessions(document_id);

CREATE TRIGGER update_student_documents_updated_at BEFORE UPDATE ON student_documents FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
