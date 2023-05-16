package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore create a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
			//return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountId   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

//var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other.
// It creates a transfer record, add account entries, and update accounts' balance within a single database transaction

func (store *Store) TrasferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		//txName := ctx.Value(txKey)

		//fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		//From the procede money
		//fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		//Account receive money
		//fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		//get account -> update its balance
		//fmt.Println(txName, "get account 1")
		/*
				account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountId)
				if err != nil {
					return err
				}


			//fmt.Println(txName, "update account 1")
			result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      arg.FromAccountId,
				Balance: account1.Balance - arg.Amount,
			})
			if err != nil {
				return err
			}
		*/

		if arg.FromAccountId < arg.ToAccountId {

			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountId, -arg.Amount, arg.ToAccountId, arg.Amount)

			/*
				result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
					ID:     arg.FromAccountId,
					Amount: -arg.Amount,
				})
				if err != nil {
					return err
				}

				result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
					ID:     arg.ToAccountId,
					Amount: arg.Amount,
				})
				if err != nil {
					return err
				}
			*/

			//fmt.Println(txName, "get account 2")
			/*
				account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountId)
				if err != nil {
					return err
				}

				//fmt.Println(txName, "update account 2")
				result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
					ID:      arg.ToAccountId,
					Balance: account2.Balance + arg.Amount,
				})
			*/

		} else {

			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountId, arg.Amount, arg.FromAccountId, -arg.Amount)

			/*
				result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
					ID:     arg.ToAccountId,
					Amount: arg.Amount,
				})
				if err != nil {
					return err
				}

				result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
					ID:     arg.FromAccountId,
					Amount: -arg.Amount,
				})
				if err != nil {
					return err
				}
			*/

		}

		return nil

	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return

}
