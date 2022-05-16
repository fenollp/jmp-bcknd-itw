package srv

import (
	"context"
	"database/sql"
	"encoding/json"

	"go.uber.org/zap"
)

// User represents a user in DB
type User struct {
	UserID              int64
	FirstName, LastName string
	Balance             int64
}

var _ json.Marshaler = (*User)(nil)

// MarshalJSON is implemented mainly to show balance as float
func (user *User) MarshalJSON() ([]byte, error) {
	type jsonUser struct {
		UserID    int64   `json:"user_id"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Balance   float64 `json:"balance"`
	}

	ju := jsonUser{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Balance:   float64(user.Balance) / 100.0,
	}

	return json.Marshal(ju)
}

func (pg *pgClient) readUsers(ctx context.Context, req *listUsersReq) (users []User, err error) {
	log := NewLogFromCtx(ctx)

	q := pg.Select("id, first_name, last_name, balance").
		From("users").
		OrderBy("id ASC").
		Limit(uint64(req.count))
	if req.fromID > 0 {
		q = q.Where("id > ?", req.fromID)
	}

	var rows *sql.Rows
	if rows, err = q.QueryContext(ctx); err != nil {
		log.Error("", zap.Error(err))
		return
	}
	defer rows.Close()

	users = make([]User, 0, req.count)
	for rows.Next() {
		var user User
		if err = rows.Scan(
			&user.UserID,
			&user.FirstName,
			&user.LastName,
			&user.Balance,
		); err != nil {
			log.Error("", zap.Error(err))
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Error("", zap.Error(err))
	}
	return
}
