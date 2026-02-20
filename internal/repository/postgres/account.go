package postgres

import (
	"database/sql"
	"vue-calc/internal/entity"
)

// AccountRepo — репозиторий для работы со счетами в PostgreSQL.
// Баланс счёта вычисляется через подзапрос (SUM всех транзакций по счёту).
type AccountRepo struct {
	db *sql.DB
}

// NewAccountRepo — конструктор репозитория счетов.
func NewAccountRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

// GetAll — получить все счета пользователя с вычисленными балансами.
func (r *AccountRepo) GetAll(userID int) ([]entity.Account, error) {
	rows, err := r.db.Query(`
		SELECT a.id, a.user_id, a.currency, a.comment, a.created_at,
		       COALESCE((SELECT SUM(t.amount) FROM transactions t WHERE t.account_id = a.id AND t.deleted_at IS NULL), 0) AS balance
		FROM accounts a
		WHERE a.user_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []entity.Account{}
	for rows.Next() {
		var a entity.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Currency, &a.Comment, &a.CreatedAt, &a.Balance); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// GetByID — получить один счёт по ID (только если принадлежит пользователю).
func (r *AccountRepo) GetByID(id, userID int) (entity.Account, error) {
	var a entity.Account
	err := r.db.QueryRow(`
		SELECT a.id, a.user_id, a.currency, a.comment, a.created_at,
		       COALESCE((SELECT SUM(t.amount) FROM transactions t WHERE t.account_id = a.id AND t.deleted_at IS NULL), 0) AS balance
		FROM accounts a
		WHERE a.id = $1 AND a.user_id = $2 AND a.deleted_at IS NULL
	`, id, userID).Scan(&a.ID, &a.UserID, &a.Currency, &a.Comment, &a.CreatedAt, &a.Balance)
	return a, err
}

// Create — создать новый счёт. Возвращает созданный счёт с присвоенным ID.
func (r *AccountRepo) Create(account entity.Account) (entity.Account, error) {
	err := r.db.QueryRow(
		"INSERT INTO accounts (currency, comment, user_id) VALUES ($1, $2, $3) RETURNING id, created_at",
		account.Currency, account.Comment, account.UserID,
	).Scan(&account.ID, &account.CreatedAt)
	return account, err
}

// Delete — мягко удалить счёт по ID (только если принадлежит пользователю).
// Также мягко удаляет все транзакции этого счёта.
func (r *AccountRepo) Delete(id, userID int) (int64, error) {
	result, err := r.db.Exec(
		"UPDATE accounts SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL",
		id, userID,
	)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected > 0 {
		_, _ = r.db.Exec(
			"UPDATE transactions SET deleted_at = NOW() WHERE account_id = $1 AND deleted_at IS NULL",
			id,
		)
	}
	return affected, nil
}

// UpdateComment — обновить комментарий счёта.
func (r *AccountRepo) UpdateComment(id, userID int, comment string) error {
	res, err := r.db.Exec(
		"UPDATE accounts SET comment = $1 WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL",
		comment, id, userID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Exists — проверить существование счёта у пользователя.
func (r *AccountRepo) Exists(id, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL)", id, userID).Scan(&exists)
	return exists, err
}
