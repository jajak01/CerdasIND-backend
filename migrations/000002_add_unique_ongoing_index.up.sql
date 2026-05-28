CREATE UNIQUE INDEX idx_unique_ongoing_ujian ON history_ujian(user_id, bundle_id) WHERE status = 'berlangsung';
