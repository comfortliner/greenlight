-- Import the data manually.

-- Table movies
INSERT INTO greenlight.dbo.movies
	(created_at,title,[year],runtime,genres,version)
VALUES
	('2022-11-12 17:04:24.09', N'TopGun', 1984, 107, N'action', 1),
	('2022-11-12 17:05:22.213', N'Moana', 2016, 107, N'animation,adventure', 1),
	('2022-11-12 17:08:15.403', N'Deadpool', 2016, 108, N'action,comedy,romance', 1),
	('2022-11-12 17:10:00.17', N'Cinderella', 1974, 120, N'fairy tale', 3),
	('2022-11-19 16:49:18.283', N'Black Panther', 2018, 234, NULL, 2),
	('2022-11-19 16:53:23.27', N'The Breakfast Club', 1985, 97, NULL, 1);
