package tools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type mockDB struct {
	mu sync.RWMutex

	// Audit trail
	transactionLogs []TransactionLog
	logMu           sync.Mutex

	// Circuit breaker for resilience
	healthStatus map[string]bool
	healthMu     sync.RWMutex

	// Performance metrics
	operationCount int64
	startTime      time.Time
}

// Mock login details database
var mockLoginDetails = map[string]LoginDetails{
	"aaron": {
		AuthToken: "1",
		Username:  "aaron",
	},
	"bryan": {
		AuthToken: "2",
		Username:  "bryan",
	},
}

// Mock coin balance database with versioning
var mockCoinDetails = map[string]CoinDetails{
	"aaron": {
		Coins:    1000,
		Username: "aaron",
		Version:  1,
	},
	"bryan": {
		Coins:    1000,
		Username: "bryan",
		Version:  1,
	},
}

func (d *mockDB) SetupDatabase() error {
	d.healthStatus = map[string]bool{
		"database":    true,
		"audit_log":   true,
		"performance": true,
	}
	d.startTime = time.Now()
	d.transactionLogs = make([]TransactionLog, 0)

	log.Info("Financial database system initialized")
	return nil
}

// Generate transaction ID
func generateTransactionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Audit logging
func (d *mockDB) logTransaction(txType, from, to string, amount int64, status string) {
	d.logMu.Lock()
	defer d.logMu.Unlock()

	txLog := TransactionLog{
		ID:        generateTransactionID(),
		Type:      txType,
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now(),
		Status:    status,
	}

	d.transactionLogs = append(d.transactionLogs, txLog)

	// Keep only last 1000 transactions (in real systems, this goes to persistent storage)
	if len(d.transactionLogs) > 1000 {
		d.transactionLogs = d.transactionLogs[len(d.transactionLogs)-1000:]
	}
}

func (d *mockDB) GetUserLoginDetails(username string) *LoginDetails {
	time.Sleep(time.Millisecond * 5)

	d.mu.RLock()
	defer d.mu.RUnlock()

	clientData, ok := mockLoginDetails[username]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) GetUserCoins(username string) *CoinDetails {
	d.mu.RLock()
	defer d.mu.RUnlock()

	clientData, ok := mockCoinDetails[username]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) AddUserCoins(username string, amount int64) *CoinDetails {
	if amount <= 0 {
		d.logTransaction("DEPOSIT", "", username, amount, "FAILED_INVALID_AMOUNT")
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	clientData, ok := mockCoinDetails[username]
	if !ok {
		d.logTransaction("DEPOSIT", "", username, amount, "FAILED_USER_NOT_FOUND")
		return nil
	}

	// Optimistic locking simulation
	clientData.Coins = clientData.Coins + amount
	clientData.Version++
	mockCoinDetails[username] = clientData

	d.logTransaction("DEPOSIT", "", username, amount, "SUCCESS")

	return &clientData
}

func (d *mockDB) WithdrawUserCoins(username string, amount int64) *CoinDetails {
	if amount <= 0 {
		d.logTransaction("WITHDRAWAL", username, "", amount, "FAILED_INVALID_AMOUNT")
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	clientData, ok := mockCoinDetails[username]
	if !ok {
		d.logTransaction("WITHDRAWAL", username, "", amount, "FAILED_USER_NOT_FOUND")
		return nil
	}

	if amount > clientData.Coins {
		d.logTransaction("WITHDRAWAL", username, "", amount, "FAILED_INSUFFICIENT_FUNDS")
		return nil
	}

	clientData.Coins = clientData.Coins - amount
	clientData.Version++
	mockCoinDetails[username] = clientData

	d.logTransaction("WITHDRAWAL", username, "", amount, "SUCCESS")

	return &clientData
}

func (d *mockDB) TransferUserCoins(from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails) {
	fromResult, toResult, err := d.TransferUserCoinsWithContext(context.Background(), from, to, amount)
	if err != nil {
		return nil, nil
	}
	return fromResult, toResult
}

// Context-aware transfer
func (d *mockDB) TransferUserCoinsWithContext(ctx context.Context, from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails, err error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_CONTEXT_CANCELLED")
		return nil, nil, ctx.Err()
	default:
	}

	if amount <= 0 {
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_INVALID_AMOUNT")
		return nil, nil, fmt.Errorf("invalid amount")
	}

	if from == to {
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_SELF_TRANSFER")
		return nil, nil, fmt.Errorf("self-transfer not allowed")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	fromData, ok := mockCoinDetails[from]
	if !ok {
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_FROM_USER_NOT_FOUND")
		return nil, nil, fmt.Errorf("sender not found")
	}

	toData, okTwo := mockCoinDetails[to]
	if !okTwo {
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_TO_USER_NOT_FOUND")
		return nil, nil, fmt.Errorf("recipient not found")
	}

	if fromData.Coins < amount {
		d.logTransaction("TRANSFER", from, to, amount, "FAILED_INSUFFICIENT_FUNDS")
		return nil, nil, fmt.Errorf("insufficient funds")
	}

	// Atomic transfer with version updates
	fromData.Coins = fromData.Coins - amount
	fromData.Version++
	mockCoinDetails[from] = fromData

	toData.Coins = toData.Coins + amount
	toData.Version++
	mockCoinDetails[to] = toData

	d.logTransaction("TRANSFER", from, to, amount, "SUCCESS")

	return &fromData, &toData, nil
}

// Financial system monitoring
func (d *mockDB) GetTransactionHistory(username string) []TransactionLog {
	d.logMu.Lock()
	defer d.logMu.Unlock()

	var userTxs []TransactionLog
	for _, tx := range d.transactionLogs {
		if tx.From == username || tx.To == username {
			userTxs = append(userTxs, tx)
		}
	}

	return userTxs
}

// System health monitoring
func (d *mockDB) GetSystemHealth() map[string]interface{} {
	d.healthMu.RLock()
	defer d.healthMu.RUnlock()

	uptime := time.Since(d.startTime)

	return map[string]interface{}{
		"status":          "healthy",
		"uptime_seconds":  uptime.Seconds(),
		"operation_count": d.operationCount,
		"components":      d.healthStatus,
		"last_check":      time.Now(),
		"version":         "1.0.0",
	}
}
