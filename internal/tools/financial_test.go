package tools

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestFinancialSystemScenarios tests advanced financial system patterns that demonstrate enterprise-level understanding of financial operations.
func TestFinancialSystemScenarios(t *testing.T) {
	t.Run("High_Frequency_Trading_Simulation", func(t *testing.T) {
		// Simulate a high-frequency trading environment
		mockCoinDetails = map[string]CoinDetails{
			"trader_1": {Coins: 100000, Username: "trader_1", Version: 1},
			"trader_2": {Coins: 100000, Username: "trader_2", Version: 1},
			"trader_3": {Coins: 100000, Username: "trader_3", Version: 1},
			"exchange": {Coins: 1000000, Username: "exchange", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup
		var successfulTrades int64
		var failedTrades int64

		numTrades := 50

		for i := 0; i < numTrades; i++ {
			wg.Add(1)
			go func(tradeID int) {
				defer wg.Done()

				traders := []string{"trader_1", "trader_2", "trader_3"}
				from := traders[rand.Intn(len(traders))]
				to := "exchange"
				amount := int64(rand.Intn(1000) + 100)

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				fromResult, toResult, err := db.TransferUserCoinsWithContext(ctx, from, to, amount)
				if err != nil || fromResult == nil || toResult == nil {
					atomic.AddInt64(&failedTrades, 1)
				} else {
					atomic.AddInt64(&successfulTrades, 1)
				}
			}(i)
		}

		wg.Wait()

		totalSuccessful := atomic.LoadInt64(&successfulTrades)
		totalFailed := atomic.LoadInt64(&failedTrades)

		t.Logf("HFT Results: %d successful, %d failed out of %d total trades",
			totalSuccessful, totalFailed, numTrades)

		// Verify money conservation
		finalExchange := db.GetUserCoins("exchange")
		finalTrader1 := db.GetUserCoins("trader_1")
		finalTrader2 := db.GetUserCoins("trader_2")
		finalTrader3 := db.GetUserCoins("trader_3")

		totalFinal := finalExchange.Coins + finalTrader1.Coins + finalTrader2.Coins + finalTrader3.Coins
		expectedTotal := int64(1300000)

		if totalFinal != expectedTotal {
			t.Errorf("MONEY CREATION/DESTRUCTION! Expected total %d, got %d", expectedTotal, totalFinal)
		}

		if totalSuccessful == 0 {
			t.Errorf("No trades succeeded - system may be deadlocked")
		}
	})

	t.Run("Bank_Run_Stress_Test", func(t *testing.T) {
		// Simulate a bank run scenario
		mockCoinDetails = map[string]CoinDetails{
			"bank":       {Coins: 500000, Username: "bank", Version: 1},
			"customer_1": {Coins: 10000, Username: "customer_1", Version: 1},
			"customer_2": {Coins: 10000, Username: "customer_2", Version: 1},
			"customer_3": {Coins: 10000, Username: "customer_3", Version: 1},
			"customer_4": {Coins: 10000, Username: "customer_4", Version: 1},
			"customer_5": {Coins: 10000, Username: "customer_5", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup
		var totalWithdrawn int64

		customers := []string{"customer_1", "customer_2", "customer_3", "customer_4", "customer_5"}

		for _, customer := range customers {
			wg.Add(1)
			go func(customerID string) {
				defer wg.Done()

				balance := db.GetUserCoins(customerID)
				if balance != nil {
					result := db.WithdrawUserCoins(customerID, balance.Coins)
					if result != nil {
						atomic.AddInt64(&totalWithdrawn, balance.Coins)
					}
				}
			}(customer)
		}

		wg.Wait()

		finalWithdrawn := atomic.LoadInt64(&totalWithdrawn)
		t.Logf("Bank run: Total withdrawn %d coins", finalWithdrawn)

		// Verify no negative balances
		for _, customer := range customers {
			balance := db.GetUserCoins(customer)
			if balance.Coins < 0 {
				t.Errorf("Customer %s has negative balance: %d", customer, balance.Coins)
			}
		}
	})

	t.Run("Payment_Processing_Workflow", func(t *testing.T) {
		// Simulate e-commerce payment processing
		mockCoinDetails = map[string]CoinDetails{
			"merchant_1":        {Coins: 50000, Username: "merchant_1", Version: 1},
			"merchant_2":        {Coins: 50000, Username: "merchant_2", Version: 1},
			"customer_a":        {Coins: 5000, Username: "customer_a", Version: 1},
			"customer_b":        {Coins: 5000, Username: "customer_b", Version: 1},
			"customer_c":        {Coins: 5000, Username: "customer_c", Version: 1},
			"payment_processor": {Coins: 0, Username: "payment_processor", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup
		var successfulPayments int64
		var failedPayments int64

		numPayments := 30

		for i := 0; i < numPayments; i++ {
			wg.Add(1)
			go func(paymentID int) {
				defer wg.Done()

				customers := []string{"customer_a", "customer_b", "customer_c"}
				merchants := []string{"merchant_1", "merchant_2"}

				customer := customers[rand.Intn(len(customers))]
				merchant := merchants[rand.Intn(len(merchants))]
				amount := int64(rand.Intn(500) + 50)

				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()

				// Two-phase payment: customer -> processor -> merchant
				_, _, err1 := db.TransferUserCoinsWithContext(ctx, customer, "payment_processor", amount)
				if err1 != nil {
					atomic.AddInt64(&failedPayments, 1)
					return
				}

				// Process fee and pay merchant
				fee := amount / 100 // 1% processing fee
				merchantAmount := amount - fee

				_, _, err2 := db.TransferUserCoinsWithContext(ctx, "payment_processor", merchant, merchantAmount)
				if err2 != nil {
					// Rollback
					db.TransferUserCoinsWithContext(ctx, "payment_processor", customer, amount)
					atomic.AddInt64(&failedPayments, 1)
					return
				}

				atomic.AddInt64(&successfulPayments, 1)
			}(i)
		}

		wg.Wait()

		finalSuccessful := atomic.LoadInt64(&successfulPayments)
		finalFailed := atomic.LoadInt64(&failedPayments)

		t.Logf("Payment processing: %d successful, %d failed out of %d total",
			finalSuccessful, finalFailed, numPayments)

		// Verify processor collected fees
		processor := db.GetUserCoins("payment_processor")
		if processor.Coins < 0 {
			t.Errorf("Payment processor has negative balance: %d", processor.Coins)
		}

		// Verify money conservation
		total := int64(0)
		for _, user := range []string{"merchant_1", "merchant_2", "customer_a", "customer_b", "customer_c", "payment_processor"} {
			balance := db.GetUserCoins(user)
			total += balance.Coins
		}

		expectedTotal := int64(115000)
		if total != expectedTotal {
			t.Errorf("Money not conserved! Expected %d, got %d", expectedTotal, total)
		}
	})

	t.Run("Deadlock_Prevention_Test", func(t *testing.T) {
		// Test circular transfer scenarios
		mockCoinDetails = map[string]CoinDetails{
			"account_a": {Coins: 10000, Username: "account_a", Version: 1},
			"account_b": {Coins: 10000, Username: "account_b", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup
		numIterations := 10

		for i := 0; i < numIterations; i++ {
			wg.Add(2)

			// Potential deadlock: A->B and B->A simultaneously
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				db.TransferUserCoinsWithContext(ctx, "account_a", "account_b", 100)
			}()

			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				db.TransferUserCoinsWithContext(ctx, "account_b", "account_a", 100)
			}()
		}

		// Test should complete without hanging
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			t.Logf("Deadlock prevention test passed")
		case <-time.After(5 * time.Second):
			t.Errorf("Potential deadlock detected - test hung")
		}

		// Verify money conservation
		balanceA := db.GetUserCoins("account_a")
		balanceB := db.GetUserCoins("account_b")
		total := balanceA.Coins + balanceB.Coins

		if total != 20000 {
			t.Errorf("Money not conserved! Expected 20000, got %d", total)
		}
	})
}

// TestComplianceAndAuditing tests features required for financial compliance
func TestComplianceAndAuditing(t *testing.T) {
	t.Run("Audit_Trail_Verification", func(t *testing.T) {
		mockCoinDetails = map[string]CoinDetails{
			"auditor": {Coins: 10000, Username: "auditor", Version: 1},
			"user_1":  {Coins: 5000, Username: "user_1", Version: 1},
			"user_2":  {Coins: 5000, Username: "user_2", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup

		// Perform auditable transactions
		transactions := []struct {
			from   string
			to     string
			amount int64
		}{
			{"auditor", "user_1", 1000},
			{"user_1", "user_2", 500},
			{"user_2", "auditor", 200},
			{"auditor", "user_2", 300},
		}

		for _, tx := range transactions {
			wg.Add(1)
			go func(from, to string, amount int64) {
				defer wg.Done()
				db.TransferUserCoins(from, to, amount)
			}(tx.from, tx.to, tx.amount)
		}

		wg.Wait()

		// Verify audit trails exist
		for _, user := range []string{"auditor", "user_1", "user_2"} {
			history := db.GetTransactionHistory(user)
			if len(history) == 0 {
				t.Errorf("No transaction history found for user %s", user)
			}
			t.Logf("User %s has %d transactions in audit trail", user, len(history))
		}

		// Verify system health
		health := db.GetSystemHealth()
		if health["status"] != "healthy" {
			t.Errorf("System health check failed: %v", health)
		}
	})

	t.Run("High_Volume_Performance_Test", func(t *testing.T) {
		// Test realistic financial system load
		mockCoinDetails = map[string]CoinDetails{
			"user_1": {Coins: 1000, Username: "user_1", Version: 1},
			"user_2": {Coins: 1000, Username: "user_2", Version: 1},
			"user_3": {Coins: 1000, Username: "user_3", Version: 1},
			"user_4": {Coins: 1000, Username: "user_4", Version: 1},
			"user_5": {Coins: 1000, Username: "user_5", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		start := time.Now()
		var wg sync.WaitGroup
		var operationCount int64

		users := []string{"user_1", "user_2", "user_3", "user_4", "user_5"}

		// 80% reads, 20% writes (realistic financial system ratio)
		for i := 0; i < 100; i++ {
			if i < 80 {
				// Read operation
				wg.Add(1)
				go func() {
					defer wg.Done()
					user := users[rand.Intn(len(users))]
					db.GetUserCoins(user)
					atomic.AddInt64(&operationCount, 1)
				}()
			} else {
				// Write operation
				wg.Add(1)
				go func() {
					defer wg.Done()
					from := users[rand.Intn(len(users))]
					to := users[rand.Intn(len(users))]
					if from != to {
						db.TransferUserCoins(from, to, 10)
					}
					atomic.AddInt64(&operationCount, 1)
				}()
			}
		}

		wg.Wait()
		duration := time.Since(start)

		finalOperationCount := atomic.LoadInt64(&operationCount)
		opsPerSecond := float64(finalOperationCount) / duration.Seconds()

		t.Logf("Performance: %d operations in %v (%.2f ops/sec)",
			finalOperationCount, duration, opsPerSecond)

		// Financial systems should handle at least 1000 ops/sec
		if opsPerSecond < 1000 {
			t.Logf("Warning: Performance below financial system standards (%.2f ops/sec)", opsPerSecond)
		}

		// Verify data integrity
		total := int64(0)
		for _, user := range users {
			balance := db.GetUserCoins(user)
			total += balance.Coins
		}

		if total != 5000 {
			t.Errorf("Data corruption detected! Expected total 5000, got %d", total)
		}
	})
}
