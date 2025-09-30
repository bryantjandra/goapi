package tools

import (
	"sync"
	"testing"
	"time"
)

// TestBasicConcurrency focuses on fundamental race condition detection and basic concurrency patterns that any financial system must handle correctly.
func TestBasicConcurrency(t *testing.T) {
	t.Run("Concurrent_Deposits", func(t *testing.T) {
		// Reset state
		mockCoinDetails = map[string]CoinDetails{
			"aaron": {Coins: 100, Username: "aaron", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup

		// Launch 3 concurrent deposits
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 10)
			}()
		}

		wg.Wait()

		// Verify result
		finalBalance := db.GetUserCoins("aaron")
		expected := int64(130) // 100 + (3 × 10)

		t.Logf("Expected: %d coins, Actually got: %d coins", expected, finalBalance.Coins)

		if finalBalance.Coins != expected {
			t.Errorf("RACE CONDITION! Expected %d, but got %d", expected, finalBalance.Coins)
		}
	})

	t.Run("Mixed_Deposits_And_Withdrawals", func(t *testing.T) {
		// Reset state
		mockCoinDetails = map[string]CoinDetails{
			"aaron": {Coins: 200, Username: "aaron", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup

		// 3 concurrent deposits
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.AddUserCoins("aaron", 20)
			}()
		}

		// 2 concurrent withdrawals
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.WithdrawUserCoins("aaron", 30)
			}()
		}

		wg.Wait()

		// Verify result
		finalBalance := db.GetUserCoins("aaron")
		expected := int64(200) // 200 + (3×20) - (2×30) = 200

		t.Logf("Expected: %d coins, Actually got: %d coins", expected, finalBalance.Coins)

		if finalBalance.Coins != expected {
			t.Errorf("RACE CONDITION! Expected %d, but got %d", expected, finalBalance.Coins)
		}
	})

	t.Run("Concurrent_Transfers", func(t *testing.T) {
		// Reset state
		mockCoinDetails = map[string]CoinDetails{
			"aaron": {Coins: 300, Username: "aaron", Version: 1},
			"bryan": {Coins: 200, Username: "bryan", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup

		// 2 transfers from Aaron to Bryan
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.TransferUserCoins("aaron", "bryan", 50)
			}()
		}

		// 2 transfers from Bryan to Aaron
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.TransferUserCoins("bryan", "aaron", 25)
			}()
		}

		wg.Wait()

		// Verify results
		aaronBalance := db.GetUserCoins("aaron")
		bryanBalance := db.GetUserCoins("bryan")

		expectedAaron := int64(250) // 300 - (2×50) + (2×25) = 250
		expectedBryan := int64(250) // 200 + (2×50) - (2×25) = 250

		t.Logf("Aaron - Expected: %d coins, Actually got: %d coins", expectedAaron, aaronBalance.Coins)
		t.Logf("Bryan - Expected: %d coins, Actually got: %d coins", expectedBryan, bryanBalance.Coins)

		if aaronBalance.Coins != expectedAaron {
			t.Errorf("RACE CONDITION! Aaron expected %d, but got %d", expectedAaron, aaronBalance.Coins)
		}

		if bryanBalance.Coins != expectedBryan {
			t.Errorf("RACE CONDITION! Bryan expected %d, but got %d", expectedBryan, bryanBalance.Coins)
		}

		// Verify money conservation
		totalCoins := aaronBalance.Coins + bryanBalance.Coins
		expectedTotal := int64(500)
		if totalCoins != expectedTotal {
			t.Errorf("MONEY CONSERVATION VIOLATED! Expected total %d, but got %d", expectedTotal, totalCoins)
		}
	})

	t.Run("Read_Write_Concurrency", func(t *testing.T) {
		// Reset state
		mockCoinDetails = map[string]CoinDetails{
			"aaron": {Coins: 150, Username: "aaron", Version: 1},
			"bryan": {Coins: 150, Username: "bryan", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		var wg sync.WaitGroup

		// Write operations
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

		wg.Add(1)
		go func() {
			defer wg.Done()
			db.TransferUserCoins("aaron", "bryan", 40)
		}()

		// Read operations (should not interfere with writes)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db.GetUserCoins("aaron")
				db.GetUserCoins("bryan")
			}()
		}

		wg.Wait()

		// Verify final state
		aaronBalance := db.GetUserCoins("aaron")
		bryanBalance := db.GetUserCoins("bryan")

		// Aaron: 150 + (2×25) - 20 - 40 = 140
		expectedAaron := int64(140)
		// Bryan: 150 + 40 = 190
		expectedBryan := int64(190)

		t.Logf("Aaron - Expected: %d coins, Actually got: %d coins", expectedAaron, aaronBalance.Coins)
		t.Logf("Bryan - Expected: %d coins, Actually got: %d coins", expectedBryan, bryanBalance.Coins)

		if aaronBalance.Coins != expectedAaron {
			t.Errorf("RACE CONDITION! Aaron expected %d, but got %d", expectedAaron, aaronBalance.Coins)
		}

		if bryanBalance.Coins != expectedBryan {
			t.Errorf("RACE CONDITION! Bryan expected %d, but got %d", expectedBryan, bryanBalance.Coins)
		}
	})
}

// TestPerformance focuses on basic performance characteristics and ensures the system can handle reasonable load efficiently.
func TestPerformance(t *testing.T) {
	t.Run("Basic_Performance_Test", func(t *testing.T) {
		// Reset state
		mockCoinDetails = map[string]CoinDetails{
			"user_1": {Coins: 1000, Username: "user_1", Version: 1},
			"user_2": {Coins: 1000, Username: "user_2", Version: 1},
		}

		database, err := NewDatabase()
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db := *database

		start := time.Now()
		var wg sync.WaitGroup

		// Simple mixed workload
		numOperations := 40

		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				switch i % 4 {
				case 0:
					db.AddUserCoins("user_1", 1)
				case 1:
					db.WithdrawUserCoins("user_2", 1)
				case 2:
					db.TransferUserCoins("user_1", "user_2", 2)
				case 3:
					db.GetUserCoins("user_1")
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Verify final state
		user1Balance := db.GetUserCoins("user_1")
		user2Balance := db.GetUserCoins("user_2")

		t.Logf("Completed %d operations in %v", numOperations, duration)
		t.Logf("User1 balance: %d, User2 balance: %d", user1Balance.Coins, user2Balance.Coins)

		// Performance should be sub-second for this load
		maxDuration := time.Second * 1
		if duration > maxDuration {
			t.Errorf("Performance issue: operations took %v, expected less than %v", duration, maxDuration)
		}

		// Verify money conservation
		total := user1Balance.Coins + user2Balance.Coins
		if total != 2000 {
			t.Errorf("Money not conserved! Expected 2000, got %d", total)
		}
	})
}

// BenchmarkBasicOperations provides performance benchmarks for individual operations
func BenchmarkBasicOperations(b *testing.B) {
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

	b.Run("GetUserCoins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.GetUserCoins("bench_user_1")
		}
	})

	b.Run("AddUserCoins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.AddUserCoins("bench_user_1", 1)
		}
	})

	b.Run("WithdrawUserCoins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.WithdrawUserCoins("bench_user_1", 1)
		}
	})

	b.Run("TransferUserCoins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				db.TransferUserCoins("bench_user_1", "bench_user_2", 1)
			} else {
				db.TransferUserCoins("bench_user_2", "bench_user_1", 1)
			}
		}
	})
}
