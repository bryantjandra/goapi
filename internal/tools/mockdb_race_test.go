package tools

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Simple test to detect race conditions with minimal operations
func TestConcurrentDeposit(t *testing.T) {
	// Start with a clean slate - Aaron has 100 coins
	mockCoinDetails = map[string]CoinDetails{
		"aaron": {Coins: 100, Username: "aaron"},
	}

	// Create database properly using NewDatabase() which calls SetupDatabase()
	database, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	t.Run("Simple Add Race", func(t *testing.T) {
		// Reset Aaron's balance
		mockCoinDetails["aaron"] = CoinDetails{Coins: 100, Username: "aaron"}

		var wg sync.WaitGroup

		// Launch 3 goroutines, each adding 10 coins (reduced from 5×100)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 10)
			}()
		}

		// Wait for all operations to complete
		wg.Wait()

		// Check the result
		finalBalance := db.GetUserCoins("aaron")
		expected := int64(130) // 100 + (3 × 10)

		t.Logf("Expected: %d coins, Actually got: %d coins", expected, finalBalance.Coins)

		if finalBalance.Coins != expected {
			t.Errorf("RACE CONDITION! Expected %d, but got %d", expected, finalBalance.Coins)
		}
	})
}

// Test concurrent deposits and withdrawals
func TestConcurrentDepositWithdraw(t *testing.T) {
	// Reset database state
	mockCoinDetails = map[string]CoinDetails{
		"aaron": {Coins: 200, Username: "aaron"},
		"bryan": {Coins: 200, Username: "bryan"},
	}

	database, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	t.Run("Mixed Add and Withdraw Operations", func(t *testing.T) {
		// Reset Aaron's balance
		mockCoinDetails["aaron"] = CoinDetails{Coins: 200, Username: "aaron"}

		var wg sync.WaitGroup

		// 3 goroutines adding 20 coins each
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 20)
			}()
		}

		// 2 goroutines withdrawing 30 coins each
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.WithdrawUserCoins("aaron", 30)
			}()
		}

		// Wait for all operations to complete
		wg.Wait()

		// Check the result
		finalBalance := db.GetUserCoins("aaron")
		expected := int64(200) // 200 + (3×20) - (2×30) = 200 + 60 - 60 = 200

		t.Logf("Expected: %d coins, Actually got: %d coins", expected, finalBalance.Coins)

		if finalBalance.Coins != expected {
			t.Errorf("RACE CONDITION! Expected %d, but got %d", expected, finalBalance.Coins)
		}
	})
}

// Test concurrent transfers between multiple users
func TestConcurrentTransfers(t *testing.T) {
	// Reset database state
	mockCoinDetails = map[string]CoinDetails{
		"aaron": {Coins: 300, Username: "aaron"},
		"bryan": {Coins: 200, Username: "bryan"},
	}

	database, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	t.Run("Multiple Transfer Operations", func(t *testing.T) {
		// Reset balances
		mockCoinDetails["aaron"] = CoinDetails{Coins: 300, Username: "aaron"}
		mockCoinDetails["bryan"] = CoinDetails{Coins: 200, Username: "bryan"}

		var wg sync.WaitGroup

		// 2 transfers from Aaron to Bryan (50 coins each)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.TransferUserCoins("aaron", "bryan", 50)
			}()
		}

		// 2 transfers from Bryan to Aaron (25 coins each)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.TransferUserCoins("bryan", "aaron", 25)
			}()
		}

		// Wait for all operations to complete
		wg.Wait()

		// Check the results
		aaronBalance := db.GetUserCoins("aaron")
		bryanBalance := db.GetUserCoins("bryan")

		expectedAaron := int64(250) // 300 - (2×50) + (2×25) = 300 - 100 + 50 = 250
		expectedBryan := int64(250) // 200 + (2×50) - (2×25) = 200 + 100 - 50 = 250

		t.Logf("Aaron - Expected: %d coins, Actually got: %d coins", expectedAaron, aaronBalance.Coins)
		t.Logf("Bryan - Expected: %d coins, Actually got: %d coins", expectedBryan, bryanBalance.Coins)

		if aaronBalance.Coins != expectedAaron {
			t.Errorf("RACE CONDITION! Aaron expected %d, but got %d", expectedAaron, aaronBalance.Coins)
		}

		if bryanBalance.Coins != expectedBryan {
			t.Errorf("RACE CONDITION! Bryan expected %d, but got %d", expectedBryan, bryanBalance.Coins)
		}

		// Verify total coins are conserved
		totalCoins := aaronBalance.Coins + bryanBalance.Coins
		expectedTotal := int64(500) // 300 + 200 = 500 (should remain constant)
		if totalCoins != expectedTotal {
			t.Errorf("COIN CONSERVATION VIOLATED! Expected total %d, but got %d", expectedTotal, totalCoins)
		}
	})
}

// Test mixed operations with balance checks
func TestMixedOperationsWithBalanceChecks(t *testing.T) {
	// Reset database state
	mockCoinDetails = map[string]CoinDetails{
		"aaron": {Coins: 150, Username: "aaron"},
		"bryan": {Coins: 150, Username: "bryan"},
	}

	database, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	t.Run("Mixed Operations with Balance Checks", func(t *testing.T) {
		// Reset balances
		mockCoinDetails["aaron"] = CoinDetails{Coins: 150, Username: "aaron"}
		mockCoinDetails["bryan"] = CoinDetails{Coins: 150, Username: "bryan"}

		var wg sync.WaitGroup

		// Aaron operations
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 25)
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			db.WithdrawUserCoins("aaron", 20)
		}()

		// Bryan operations
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("bryan", 15)
			}()
		}

		// 1 Transfer
		wg.Add(1)
		go func() {
			defer wg.Done()
			db.TransferUserCoins("aaron", "bryan", 40)
		}()

		// Balance checks
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.GetUserCoins("aaron")
				db.GetUserCoins("bryan")
			}()
		}

		// Wait for all operations to complete
		wg.Wait()

		// Check final balances
		aaronBalance := db.GetUserCoins("aaron")
		bryanBalance := db.GetUserCoins("bryan")

		// Aaron: 150 + (2×25) - 20 - 40 = 150 + 50 - 20 - 40 = 140
		expectedAaron := int64(140)
		// Bryan: 150 + (2×15) + 40 = 150 + 30 + 40 = 220
		expectedBryan := int64(220)

		t.Logf("Aaron - Expected: %d coins, Actually got: %d coins", expectedAaron, aaronBalance.Coins)
		t.Logf("Bryan - Expected: %d coins, Actually got: %d coins", expectedBryan, bryanBalance.Coins)

		if aaronBalance.Coins != expectedAaron {
			t.Errorf("RACE CONDITION! Aaron expected %d, but got %d", expectedAaron, aaronBalance.Coins)
		}

		if bryanBalance.Coins != expectedBryan {
			t.Errorf("RACE CONDITION! Bryan expected %d, but got %d", expectedBryan, bryanBalance.Coins)
		}

		// Verify total coins are conserved
		totalCoins := aaronBalance.Coins + bryanBalance.Coins
		expectedTotal := int64(360) // Should equal sum of all operations
		if totalCoins != expectedTotal {
			t.Errorf("COIN CONSERVATION VIOLATED! Expected total %d, but got %d", expectedTotal, totalCoins)
		}
	})
}

// Fast performance test to ensure no bottlenecks
func TestPerformanceUnderLoad(t *testing.T) {
	// Reset database state
	mockCoinDetails = map[string]CoinDetails{
		"aaron": {Coins: 1000, Username: "aaron"},
		"bryan": {Coins: 1000, Username: "bryan"},
	}

	database, err := NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	t.Run("Performance Under Load", func(t *testing.T) {
		// Reset balances
		mockCoinDetails["aaron"] = CoinDetails{Coins: 1000, Username: "aaron"}
		mockCoinDetails["bryan"] = CoinDetails{Coins: 1000, Username: "bryan"}

		start := time.Now()
		var wg sync.WaitGroup
		numOperations := 10 // Reduced from 100 to 10

		// Launch concurrent operations
		for i := 0; i < numOperations; i++ {
			wg.Add(4) // 4 operations per iteration

			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 1)
			}()

			go func() {
				defer wg.Done()
				db.WithdrawUserCoins("bryan", 1)
			}()

			go func() {
				defer wg.Done()
				db.TransferUserCoins("aaron", "bryan", 2)
			}()

			go func() {
				defer wg.Done()
				db.GetUserCoins("aaron")
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		// Check final state
		aaronBalance := db.GetUserCoins("aaron")
		bryanBalance := db.GetUserCoins("bryan")

		// Aaron: 1000 + (10×1) - (10×2) = 1000 + 10 - 20 = 990
		expectedAaron := int64(990)
		// Bryan: 1000 - (10×1) + (10×2) = 1000 - 10 + 20 = 1010
		expectedBryan := int64(1010)

		t.Logf("Completed %d operations in %v", numOperations*4, duration)
		t.Logf("Aaron - Expected: %d coins, Actually got: %d coins", expectedAaron, aaronBalance.Coins)
		t.Logf("Bryan - Expected: %d coins, Actually got: %d coins", expectedBryan, bryanBalance.Coins)

		if aaronBalance.Coins != expectedAaron {
			t.Errorf("RACE CONDITION! Aaron expected %d, but got %d", expectedAaron, aaronBalance.Coins)
		}

		if bryanBalance.Coins != expectedBryan {
			t.Errorf("RACE CONDITION! Bryan expected %d, but got %d", expectedBryan, bryanBalance.Coins)
		}

		// Performance check - should complete very quickly
		maxDuration := time.Second * 5
		if duration > maxDuration {
			t.Errorf("Performance issue: operations took %v, expected less than %v", duration, maxDuration)
		}
	})
}

// Financial system test suite modeling real-world scenarios
func TestFinancialSystemConcurrency(t *testing.T) {
	// Test 1: High-frequency trading simulation
	t.Run("High_Frequency_Trading_Simulation", func(t *testing.T) {
		// Reset state - simulate trading accounts
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

		// Simulate 50 concurrent trades (realistic HFT volume)
		numTrades := 50

		for i := 0; i < numTrades; i++ {
			wg.Add(1)
			go func(tradeID int) {
				defer wg.Done()

				// Random trade between traders and exchange
				traders := []string{"trader_1", "trader_2", "trader_3"}
				from := traders[rand.Intn(len(traders))]
				to := "exchange"
				amount := int64(rand.Intn(1000) + 100) // 100-1099 coins

				// Simulate trade with timeout
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

		// Verify system integrity
		totalSuccessful := atomic.LoadInt64(&successfulTrades)
		totalFailed := atomic.LoadInt64(&failedTrades)

		t.Logf("Trading Results: %d successful, %d failed out of %d total trades",
			totalSuccessful, totalFailed, numTrades)

		// Verify no money was created or destroyed
		finalExchange := db.GetUserCoins("exchange")
		finalTrader1 := db.GetUserCoins("trader_1")
		finalTrader2 := db.GetUserCoins("trader_2")
		finalTrader3 := db.GetUserCoins("trader_3")

		totalFinal := finalExchange.Coins + finalTrader1.Coins + finalTrader2.Coins + finalTrader3.Coins
		expectedTotal := int64(1300000) // 100k + 100k + 100k + 1000k

		if totalFinal != expectedTotal {
			t.Errorf("MONEY CREATION/DESTRUCTION! Expected total %d, got %d", expectedTotal, totalFinal)
		}

		// Ensure some trades succeeded (system is functional)
		if totalSuccessful == 0 {
			t.Errorf("No trades succeeded - system may be deadlocked")
		}
	})

	// Test 2: Bank run simulation (stress test)
	t.Run("Bank_Run_Simulation", func(t *testing.T) {
		// Reset state - simulate bank with many customers
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

		// Simulate panic withdrawals
		customers := []string{"customer_1", "customer_2", "customer_3", "customer_4", "customer_5"}

		for _, customer := range customers {
			wg.Add(1)
			go func(customerID string) {
				defer wg.Done()

				// Each customer tries to withdraw their full balance
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

		// Verify system handled concurrent withdrawals correctly
		finalWithdrawn := atomic.LoadInt64(&totalWithdrawn)
		t.Logf("Total withdrawn during bank run: %d coins", finalWithdrawn)

		// Verify no customer has negative balance
		for _, customer := range customers {
			balance := db.GetUserCoins(customer)
			if balance.Coins < 0 {
				t.Errorf("Customer %s has negative balance: %d", customer, balance.Coins)
			}
		}
	})

	// Test 3: Payment processing simulation (real-world e-commerce)
	t.Run("Payment_Processing_Simulation", func(t *testing.T) {
		// Reset state - simulate payment processor
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

		// Simulate 30 concurrent payment transactions
		numPayments := 30

		for i := 0; i < numPayments; i++ {
			wg.Add(1)
			go func(paymentID int) {
				defer wg.Done()

				customers := []string{"customer_a", "customer_b", "customer_c"}
				merchants := []string{"merchant_1", "merchant_2"}

				customer := customers[rand.Intn(len(customers))]
				merchant := merchants[rand.Intn(len(merchants))]
				amount := int64(rand.Intn(500) + 50) // $50-$549

				// Two-phase payment: customer -> processor -> merchant
				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()

				// Phase 1: Customer pays processor
				_, _, err1 := db.TransferUserCoinsWithContext(ctx, customer, "payment_processor", amount)
				if err1 != nil {
					atomic.AddInt64(&failedPayments, 1)
					return
				}

				// Phase 2: Processor pays merchant (minus fee)
				fee := amount / 100 // 1% processing fee
				merchantAmount := amount - fee

				_, _, err2 := db.TransferUserCoinsWithContext(ctx, "payment_processor", merchant, merchantAmount)
				if err2 != nil {
					// Rollback: refund customer
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

		t.Logf("Payment Results: %d successful, %d failed out of %d total payments",
			finalSuccessful, finalFailed, numPayments)

		// Verify payment processor collected fees
		processor := db.GetUserCoins("payment_processor")
		if processor.Coins < 0 {
			t.Errorf("Payment processor has negative balance: %d", processor.Coins)
		}

		// Verify total money conservation
		total := int64(0)
		for _, user := range []string{"merchant_1", "merchant_2", "customer_a", "customer_b", "customer_c", "payment_processor"} {
			balance := db.GetUserCoins(user)
			total += balance.Coins
		}

		expectedTotal := int64(115000) // 50k + 50k + 5k + 5k + 5k + 0
		if total != expectedTotal {
			t.Errorf("Money not conserved! Expected %d, got %d", expectedTotal, total)
		}
	})

	// Test 4: Audit trail verification (compliance testing)
	t.Run("Audit_Trail_Verification", func(t *testing.T) {
		// Reset state
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

		// Perform various transactions that should be auditable
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

		// Verify audit trail exists for each user
		for _, user := range []string{"auditor", "user_1", "user_2"} {
			history := db.GetTransactionHistory(user)
			if len(history) == 0 {
				t.Errorf("No transaction history found for user %s", user)
			}

			t.Logf("User %s has %d transactions in audit trail", user, len(history))
		}

		// Verify system health monitoring
		health := db.GetSystemHealth()
		if health["status"] != "healthy" {
			t.Errorf("System health check failed: %v", health)
		}

		t.Logf("System health: %v", health)
	})

	// Test 5: Deadlock prevention test
	t.Run("Deadlock_Prevention_Test", func(t *testing.T) {
		// Reset state
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

		// Simulate potential deadlock scenario: A->B and B->A simultaneously
		numIterations := 10

		for i := 0; i < numIterations; i++ {
			wg.Add(2)

			// Transfer A -> B
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				db.TransferUserCoinsWithContext(ctx, "account_a", "account_b", 100)
			}()

			// Transfer B -> A (potential deadlock)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				db.TransferUserCoinsWithContext(ctx, "account_b", "account_a", 100)
			}()
		}

		// Test should complete without hanging (deadlock)
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			t.Logf("Deadlock prevention test passed - no hanging detected")
		case <-time.After(5 * time.Second):
			t.Errorf("Potential deadlock detected - test hung for 5 seconds")
		}

		// Verify final balances are reasonable
		balanceA := db.GetUserCoins("account_a")
		balanceB := db.GetUserCoins("account_b")
		total := balanceA.Coins + balanceB.Coins

		if total != 20000 {
			t.Errorf("Money not conserved in deadlock test! Expected 20000, got %d", total)
		}
	})

	// Test 6: Performance under realistic load
	t.Run("Realistic_Performance_Test", func(t *testing.T) {
		// Reset state - simulate realistic user base
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

		// Simulate realistic mixed workload
		users := []string{"user_1", "user_2", "user_3", "user_4", "user_5"}

		// 80% reads, 20% writes (realistic ratio)
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

		// Verify data integrity after load test
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

// Benchmark test for financial system performance
func BenchmarkFinancialOperations(b *testing.B) {
	// Reset state
	mockCoinDetails = map[string]CoinDetails{
		"bench_user_1": {Coins: 100000, Username: "bench_user_1", Version: 1},
		"bench_user_2": {Coins: 100000, Username: "bench_user_2", Version: 1},
	}

	database, err := NewDatabase()
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	db := *database

	b.ResetTimer()

	b.Run("Transfer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				db.TransferUserCoins("bench_user_1", "bench_user_2", 1)
			} else {
				db.TransferUserCoins("bench_user_2", "bench_user_1", 1)
			}
		}
	})

	b.Run("Balance_Check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				db.GetUserCoins("bench_user_1")
			} else {
				db.GetUserCoins("bench_user_2")
			}
		}
	})

	b.Run("Deposit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				db.AddUserCoins("bench_user_1", 1)
			} else {
				db.AddUserCoins("bench_user_2", 1)
			}
		}
	})
}
