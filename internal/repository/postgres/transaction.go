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
	rows, err := r.db.Query(`
		SELECT t.id, t.account_id, t.amount, t.comment, t.category_id, COALESCE(c.name, ''), t.created_at
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.account_id = $1
		ORDER BY t.created_at DESC`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []entity.Transaction{}
	for rows.Next() {
		var t entity.Transaction
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Comment, &t.CategoryID, &t.Category, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

// Delete — удалить транзакцию по ID и account_id.
func (r *TransactionRepo) Delete(id, accountID int) error {
	res, err := r.db.Exec(
		"DELETE FROM transactions WHERE id = $1 AND account_id = $2",
		id, accountID,
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

// Update — обновить транзакцию по ID и account_id.
func (r *TransactionRepo) Update(id, accountID int, transaction entity.Transaction) (entity.Transaction, error) {
	err := r.db.QueryRow(`
		UPDATE transactions SET amount=$1, comment=$2, category_id=$3, created_at=$4
		WHERE id=$5 AND account_id=$6
		RETURNING id, account_id, amount, comment, category_id, created_at`,
		transaction.Amount, transaction.Comment, transaction.CategoryID, transaction.CreatedAt,
		id, accountID,
	).Scan(&transaction.ID, &transaction.AccountID, &transaction.Amount, &transaction.Comment, &transaction.CategoryID, &transaction.CreatedAt)
	if err != nil {
		return entity.Transaction{}, err
	}

	// Fetch category name
	if transaction.CategoryID != nil {
		_ = r.db.QueryRow("SELECT name FROM categories WHERE id = $1", *transaction.CategoryID).Scan(&transaction.Category)
	}

	return transaction, nil
}

// Create — создать новую транзакцию (операцию) по счёту.
func (r *TransactionRepo) Create(transaction entity.Transaction) (entity.Transaction, error) {
	if transaction.CreatedAt != "" {
		err := r.db.QueryRow(
			"INSERT INTO transactions (account_id, amount, comment, category_id, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
			transaction.AccountID, transaction.Amount, transaction.Comment, transaction.CategoryID, transaction.CreatedAt,
		).Scan(&transaction.ID, &transaction.CreatedAt)
		return transaction, err
	}
	err := r.db.QueryRow(
		"INSERT INTO transactions (account_id, amount, comment, category_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
		transaction.AccountID, transaction.Amount, transaction.Comment, transaction.CategoryID,
	).Scan(&transaction.ID, &transaction.CreatedAt)
	return transaction, err
}
