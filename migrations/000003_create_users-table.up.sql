CREATE TABLE [users] (
  [id] bigint PRIMARY KEY IDENTITY(1, 1),
  [created_at] datetime NOT NULL DEFAULT (getdate()),
  [name] nvarchar(255) NOT NULL,
  [email] nvarchar(255) UNIQUE NOT NULL,
  [password_hash] char(60) NOT NULL,
  [activated] bit NOT NULL,
  [version] int NOT NULL DEFAULT 1
)
