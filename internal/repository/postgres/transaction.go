package postgres

import (
	"database/sql"
	"vue-calc/internal/entity"
)

// TransactionRepo — репозиторий для работы с операциями (транзакциями) в PostgreSQL.
type TransactionRepo struct {
	db *sql.DB
}

// NewTransactionRepo — конструктор репозитория транзакций.
func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// GetByAccountID — получить все транзакции по счёту, новые сверху (ORDER BY created_at DESC).
func (r *TransactionRepo) GetByAccountID(accountID int) ([]entity.Transaction, error) {
	rows, err := r.db.Query(
		"SELECT id, account_id, amount, comment, created_at FROM transactions WHERE account_id = $1 ORDER BY created_at DESC",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []entity.Transaction{}
	for rows.Next() {
		var t entity.Transaction
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Comment, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

// Create — создать новую транзакцию (операцию) по счёту.
func (r *TransactionRepo) Create(transaction entity.Transaction) (entity.Transaction, error) {
	err := r.db.QueryRow(
		"INSERT INTO transactions (account_id, amount, comment) VALUES ($1, $2, $3) RETURNING id, created_at",
		transaction.AccountID, transaction.Amount, transaction.Comment,
	).Scan(&transaction.ID, &transaction.CreatedAt)
	return transaction, err
}
