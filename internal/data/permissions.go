package data

import (
	"context"
	"database/sql"
	"time"
)

// **********************
// * Model Definition
// **********************

// Define a Permissions slice, which we will use to hold the permission codes for a single user.
type Permissions []string

// Add a helper method to check whether the Permissions slice contains a specific permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

// Define a PermissionModel struct type which wraps a sql.DB connection pool.
type PermissionModel struct {
	DB *sql.DB
}

// **********************
// * Data Validation
// **********************

// **********************
// * Data Manipulation
// **********************

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
	SELECT permissions.code
	FROM permissions
	INNER JOIN userpermissions ON userpermissions.permission_id = permissions.id
	INNER JOIN users ON userpermissions.user_id = users.id
	WHERE users.id = @p1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, code string) error {
	query := `
	INSERT INTO userpermissions
	SELECT @p1, permissions.id FROM permissions WHERE code in (@p2);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, code)
	return err
}
