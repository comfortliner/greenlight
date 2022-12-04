CREATE TABLE [usertokens] (
  [hash] varbinary(8000) PRIMARY KEY,
  [user_id] bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  [expiry] datetimeoffset NOT NULL,
  [scope] nvarchar(255) NOT NULL
)
