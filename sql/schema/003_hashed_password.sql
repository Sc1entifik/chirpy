-- +goose up
ALTER TABLE users
ADD hashed_password TEXT NOT Null
  CONSTRAINT hashed_password DEFAULT 'unset';


-- +goose down
ALTER TABLE users
DROP COLUMN hashed_password;
