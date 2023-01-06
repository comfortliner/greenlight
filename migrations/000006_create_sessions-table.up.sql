CREATE TABLE [sessions] (
	[token] CHAR(43) PRIMARY KEY,
	[data] VARBINARY(MAX) NOT NULL,
	[expiry] DATETIME2(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
