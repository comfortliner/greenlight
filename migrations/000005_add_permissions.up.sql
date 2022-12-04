CREATE TABLE [permissions] (
  [id] bigint PRIMARY KEY IDENTITY(1, 1),
  [code] nvarchar(255) NOT NULL
);

CREATE TABLE [userpermissions] (
  [user_id] bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  [permission_id] bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
  CONSTRAINT PK_userpermissions PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES
  ('movies:read'),
  ('movies:write');
