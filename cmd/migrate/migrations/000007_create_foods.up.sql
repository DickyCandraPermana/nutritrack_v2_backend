CREATE TABLE IF NOT EXISTS foods(
  id bigserial PRIMARY KEY,
  name varchar(255),
  descrption text,
  nutrition JSONB,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
)