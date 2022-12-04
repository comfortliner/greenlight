CREATE TABLE [movies] (
  [id] bigint PRIMARY KEY IDENTITY(1, 1),
  [created_at] datetime NOT NULL DEFAULT (getdate()),
  [title] nvarchar(255) NOT NULL,
  [year] int NOT NULL,
  [runtime] int NOT NULL,
  [genres] nvarchar(max),
  [version] int NOT NULL DEFAULT 1
)
