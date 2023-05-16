package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	Account1 := createRandomAccount(t)
	Account2 := createRandomAccount(t)
	fmt.Println(">> before", Account1.Balance, Account2.Balance)

	//run n current transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		//txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.Background()
			//ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TrasferTx(ctx, TransferTxParams{
				FromAccountId: Account1.ID,
				ToAccountId:   Account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, Account1.ID, transfer.FromAccountID)
		require.Equal(t, Account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, Account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, Account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, Account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, Account2.ID, toAccount.ID)

		//Check accounts balance
		fmt.Println(">> tx ln92", fromAccount.Balance, toAccount.Balance)
		diff1 := Account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - Account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amounnt, ....., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true

	}

	// check the final updated balances

	updateAccount1, err := testQueries.GetAccount(context.Background(), Account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), Account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after", updateAccount1.Balance, updateAccount2.Balance)
	require.Equal(t, Account1.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, Account2.Balance+int64(n)*amount, updateAccount2.Balance)
}

func TestTransferDeadlockTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before", account1.Balance, account2.Balance)

	//run n current transfer transactions
	n := 10
	amount := int64(10)
	errs := make(chan error)
	//results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		//txName := fmt.Sprintf("tx %d", i+1)
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {

			//ctx := context.WithValue(context.Background(), txKey, txName)
			_, err := store.TrasferTx(context.Background(), TransferTxParams{
				FromAccountId: fromAccountID,
				ToAccountId:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	// check results
	//existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		/*
			//check transfer
			transfer := result.Transfer
			require.NotEmpty(t, transfer)
			require.Equal(t, Account1.ID, transfer.FromAccountID)
			require.Equal(t, Account2.ID, transfer.ToAccountID)
			require.Equal(t, amount, transfer.Amount)
			require.NotZero(t, transfer.ID)
			require.NotZero(t, transfer.CreatedAt)

			_, err = store.GetTransfer(context.Background(), transfer.ID)
			require.NoError(t, err)

			//check entries
			fromEntry := result.FromEntry
			require.NotEmpty(t, fromEntry)
			require.Equal(t, Account1.ID, fromEntry.AccountID)
			require.Equal(t, -amount, fromEntry.Amount)
			require.NotZero(t, fromEntry.ID)
			require.NotZero(t, fromEntry.CreatedAt)

			_, err = store.GetEntry(context.Background(), fromEntry.ID)
			require.NoError(t, err)

			toEntry := result.ToEntry
			require.NotEmpty(t, toEntry)
			require.Equal(t, Account2.ID, toEntry.AccountID)
			require.Equal(t, amount, toEntry.Amount)
			require.NotZero(t, toEntry.ID)
			require.NotZero(t, toEntry.CreatedAt)

			_, err = store.GetEntry(context.Background(), toEntry.ID)
			require.NoError(t, err)

			//check accounts
			fromAccount := result.FromAccount
			require.NotEmpty(t, fromAccount)
			require.Equal(t, Account1.ID, fromAccount.ID)

			toAccount := result.ToAccount
			require.NotEmpty(t, toAccount)
			require.Equal(t, Account2.ID, toAccount.ID)

			//Check accounts balance
			fmt.Println(">> tx ln92", fromAccount.Balance, toAccount.Balance)
			diff1 := Account1.Balance - fromAccount.Balance
			diff2 := toAccount.Balance - Account2.Balance
			require.Equal(t, diff1, diff2)
			require.True(t, diff1 > 0)
			require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amounnt, ....., n * amount

			k := int(diff1 / amount)
			require.True(t, k >= 1 && k <= n)
			require.NotContains(t, existed, k)
			existed[k] = true
		*/
	}

	// check the final updated balances

	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after", updateAccount1.Balance, updateAccount2.Balance)
	require.Equal(t, account1.Balance, updateAccount1.Balance)
	require.Equal(t, account2.Balance, updateAccount2.Balance)

}
