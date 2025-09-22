# 📄 PRD: Phase 1 — Withdrawals and Transfers for GoLedger

## 🎯 Objective

Expand the current system by introducing **withdrawals** and **transfers** alongside deposits. This allows users to not only add coins but also **subtract** them (withdraw) and **move coins between accounts** (transfer).

---

## 🛠 Requirements

### 1. Withdraw Operation

* Add support for a `POST /account/coins/withdraw` endpoint.
* Accepts `username` and `amount`.
* Business rules:

  * If user not found → return error.
  * If balance is less than amount → return error (“insufficient funds”).
  * Otherwise subtract from user’s balance.

### 2. Transfer Operation

* Add support for a `POST /account/coins/transfer` endpoint.
* Accepts `from`, `to`, and `amount`.
* Business rules:

  * If either user not found → return error.
  * If sender’s balance < amount → return error.
  * Otherwise subtract from `from`, add to `to`.

### 3. Responses

* Both endpoints should return:

  * Success code (200).
  * Message string (“withdrawal successful”, “transfer successful”).
  * Updated balances for the relevant account(s).

---

## 📦 Deliverables

* Extend `DatabaseInterface` with methods:

  * `WithdrawUserCoins(username string, amount int64) *CoinDetails`
  * `TransferUserCoins(from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails)`
* Implement these methods in `mockdb.go`.
* Add request/response structs in `api/api.go` for withdrawals and transfers.
* Create new handlers in `internal/handlers/` for `WithdrawCoins` and `TransferCoins`.
* Register new routes in `internal/handlers/api.go`:

  * `router.Post("/coins/withdraw", WithdrawCoins)`
  * `router.Post("/coins/transfer", TransferCoins)`

---

## ✅ Success Criteria

* Users can successfully withdraw coins if they have enough balance.
* Users can successfully transfer coins between accounts if the sender has enough balance.
* Errors are returned properly when accounts are missing or balances are insufficient.
* API responses include updated balances for clarity.

---

## 📚 Learning Mode (Important Instruction for Cursor)

* Cursor should **only assist and guide**:

  * Tell me which files need editing.
  * Explain **why** the edits are necessary.
  * Clarify concepts about Go and backend design.
* Cursor should **not output complete code**. I (the user) will write the code myself to learn Go inside and out.
