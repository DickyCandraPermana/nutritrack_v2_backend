-- Hapus constraint lama jika ada
ALTER TABLE users DROP CONSTRAINT users_email_key;
ALTER TABLE users DROP CONSTRAINT users_username_key;

-- Buat index unik hanya untuk yang deleted_at-nya NULL
CREATE UNIQUE INDEX users_email_unique_active 
ON users (email) 
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX users_username_unique_active 
ON users (username) 
WHERE deleted_at IS NULL;