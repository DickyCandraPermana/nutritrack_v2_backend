ALTER TABLE foods 
  ALTER COLUMN name TYPE VARCHAR(255),
  ALTER COLUMN name SET NOT NULL;

ALTER TABLE foods 
  RENAME COLUMN descrption TO description;

ALTER TABLE foods
  RENAME COLUMN nutrition TO nutrients;