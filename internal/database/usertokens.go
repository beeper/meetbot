package database

import (
	"context"

	"maunium.net/go/mautrix/id"
)

func (db *Database) GetUserRefreshToken(ctx context.Context, userID id.UserID) (token string, err error) {
	err = db.DB.
		QueryRowContext(ctx, "SELECT token FROM user_refresh_tokens WHERE user_id = $1", userID).
		Scan(&token)
	return
}

func (db *Database) SetUserRefreshToken(ctx context.Context, userID id.UserID, token string) (err error) {
	q := `
		INSERT INTO user_refresh_tokens (user_id, token)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
			SET token = $2
	`
	_, err = db.DB.ExecContext(ctx, q, userID, token)
	return
}
