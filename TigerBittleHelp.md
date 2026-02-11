# TigerBeetle Transaction List Queries

## Overview

TigerBeetle provides multiple ways to query transfers (transactions) for a given account:

1. **`get_account_transfers`** - Query transfers by Account ID (debits or credits)
2. **`query_transfers`** - Query transfers by multiple filter criteria
3. **`lookup_transfers`** - Fetch specific transfers by ID

## 1. Query Transfers by Account ID

### Go Code Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    tb "github.com/tigerbeetle/tigerbeetle-go"
    "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func main() {
    // Connect to TigerBeetle
    client, err := tb.NewClient(0, []string{"localhost:3000"}, 32)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Example Account ID to query
    accountID := types.HexStringToUint128("0123456789abcdef0123456789abcdef")

    // Create filter to get transfers for this account
    filter := types.AccountFilter{
        AccountID:    accountID,
        TimestampMin: 0,                    // From beginning of time
        TimestampMax: 0,                    // To now (0 means no upper limit)
        Limit:        100,                  // Max 100 results
        Flags:        types.AccountFilterFlags{
            Debits:   true,                 // Include debits (money leaving account)
            Credits:  true,                 // Include credits (money entering account)
            Reversed: false,                // Chronological order (oldest first)
        }.ToUint32(),
    }

    // Query transfers
    transfers, err := client.GetAccountTransfers(filter)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d transfers for account %s\n", len(transfers), accountID.String())
    
    // Convert to JSON
    jsonData, _ := json.MarshalIndent(transfers, "", "  ")
    fmt.Println(string(jsonData))
}
```

### Alternative: Query by Account Code

```go
// Query transfers by Account Code using query_transfers
func QueryByAccountCode(client *tb.Client, code uint16, ledger uint32) ([]types.Transfer, error) {
    filter := types.QueryFilter{
        UserData128:  types.ToUint128(0),   // Not filtering by this
        UserData64:   0,                    // Not filtering by this
        UserData32:   0,                    // Not filtering by this
        Ledger:       ledger,               // Filter by ledger
        Code:         code,                 // Filter by transfer code
        TimestampMin: 0,
        TimestampMax: 0,
        Limit:        100,
        Flags: types.QueryFilterFlags{
            Reversed: false,                // Chronological order
        }.ToUint32(),
    }

    transfers, err := client.QueryTransfers(filter)
    return transfers, err
}
```

## 2. Go Struct Definitions

```go
package models

import "github.com/tigerbeetle/tigerbeetle-go/pkg/types"

// TransferResponse represents a transfer in API responses
type TransferResponse struct {
    ID              string `json:"id"`
    DebitAccountID  string `json:"debit_account_id"`
    CreditAccountID string `json:"credit_account_id"`
    Amount          string `json:"amount"`
    PendingID       string `json:"pending_id,omitempty"`
    UserData128     string `json:"user_data_128,omitempty"`
    UserData64      uint64 `json:"user_data_64,omitempty"`
    UserData32      uint32 `json:"user_data_32,omitempty"`
    Timeout         uint32 `json:"timeout,omitempty"`
    Ledger          uint32 `json:"ledger"`
    Code            uint16 `json:"code"`
    Flags           string `json:"flags"`
    Timestamp       uint64 `json:"timestamp"`
}

// TransferFilter for querying transfers
type TransferFilter struct {
    AccountID    string `json:"account_id,omitempty"`
    Code         uint16 `json:"code,omitempty"`
    Ledger       uint32 `json:"ledger,omitempty"`
    TimestampMin uint64 `json:"timestamp_min,omitempty"`
    TimestampMax uint64 `json:"timestamp_max,omitempty"`
    Limit        uint32 `json:"limit"`
    Debits       bool   `json:"debits"`
    Credits      bool   `json:"credits"`
    Reversed     bool   `json:"reversed"`
}

// TransferListResponse for API responses
type TransferListResponse struct {
    Transfers  []TransferResponse `json:"transfers"`
    TotalCount int                `json:"total_count"`
    HasMore    bool               `json:"has_more"`
}

// Helper function to convert TigerBeetle Transfer to API response
func ToTransferResponse(t types.Transfer) TransferResponse {
    flags := []string{}
    transferFlags := t.TransferFlags()
    
    if transferFlags.Linked {
        flags = append(flags, "linked")
    }
    if transferFlags.Pending {
        flags = append(flags, "pending")
    }
    if transferFlags.PostPendingTransfer {
        flags = append(flags, "post_pending_transfer")
    }
    if transferFlags.VoidPendingTransfer {
        flags = append(flags, "void_pending_transfer")
    }
    
    flagsStr := ""
    if len(flags) > 0 {
        for i, flag := range flags {
            if i > 0 {
                flagsStr += ","
            }
            flagsStr += flag
        }
    }
    
    return TransferResponse{
        ID:              t.ID.String(),
        DebitAccountID:  t.DebitAccountID.String(),
        CreditAccountID: t.CreditAccountID.String(),
        Amount:          t.Amount.String(),
        PendingID:       t.PendingID.String(),
        UserData128:     t.UserData128.String(),
        UserData64:      t.UserData64,
        UserData32:      t.UserData32,
        Timeout:         t.Timeout,
        Ledger:          t.Ledger,
        Code:            t.Code,
        Flags:           flagsStr,
        Timestamp:       t.Timestamp,
    }
}
```

## 3. Example JSON Response

### Single Transfer Query Result

```json
{
  "transfers": [
    {
      "id": "01936d89a1234567890abcdef0123456",
      "debit_account_id": "01936d89a0000000000000000000001",
      "credit_account_id": "01936d89a0000000000000000000002",
      "amount": "10000",
      "pending_id": "0",
      "user_data_128": "550e8400e29b41d4a716446655440000",
      "user_data_64": 1234567890,
      "user_data_32": 100,
      "timeout": 0,
      "ledger": 1,
      "code": 718,
      "flags": "",
      "timestamp": 1738665600000000000
    },
    {
      "id": "01936d89a1234567890abcdef0123457",
      "debit_account_id": "01936d89a0000000000000000000002",
      "credit_account_id": "01936d89a0000000000000000000003",
      "amount": "5000",
      "pending_id": "0",
      "user_data_128": "550e8400e29b41d4a716446655440001",
      "user_data_64": 1234567891,
      "user_data_32": 101,
      "timeout": 0,
      "ledger": 1,
      "code": 718,
      "flags": "",
      "timestamp": 1738665601000000000
    },
    {
      "id": "01936d89a1234567890abcdef0123458",
      "debit_account_id": "01936d89a0000000000000000000001",
      "credit_account_id": "01936d89a0000000000000000000004",
      "amount": "2500",
      "pending_id": "01936d89a1234567890abcdef0123456",
      "user_data_128": "0",
      "user_data_64": 0,
      "user_data_32": 0,
      "timeout": 3600,
      "ledger": 1,
      "code": 718,
      "flags": "pending",
      "timestamp": 1738665602000000000
    }
  ],
  "total_count": 3,
  "has_more": false
}
```

### Transfer List with Pagination

```json
{
  "transfers": [
    {
      "id": "01936d89a1234567890abcdef0123456",
      "debit_account_id": "01936d89a0000000000000000000001",
      "credit_account_id": "01936d89a0000000000000000000002",
      "amount": "10000",
      "ledger": 1,
      "code": 718,
      "flags": "",
      "timestamp": 1738665600000000000
    }
  ],
  "total_count": 150,
  "has_more": true,
  "next_timestamp": 1738665601000000000
}
```

## 4. gRPC Protocol Buffers Definition

```protobuf
syntax = "proto3";

package tigerbeetle.v1;

option go_package = "github.com/yourcompany/yourapp/api/tigerbeetle/v1;tigerbeettlev1";

import "google/protobuf/timestamp.proto";

// Service definition
service TigerBeetleService {
  // Get transfers for a specific account
  rpc GetAccountTransfers(GetAccountTransfersRequest) returns (GetAccountTransfersResponse);
  
  // Query transfers by various filters
  rpc QueryTransfers(QueryTransfersRequest) returns (QueryTransfersResponse);
  
  // Lookup specific transfers by ID
  rpc LookupTransfers(LookupTransfersRequest) returns (LookupTransfersResponse);
}

// Messages

message Uint128 {
  // 128-bit unsigned integer represented as bytes (little-endian)
  bytes value = 1; // 16 bytes
}

message Transfer {
  Uint128 id = 1;
  Uint128 debit_account_id = 2;
  Uint128 credit_account_id = 3;
  Uint128 amount = 4;
  Uint128 pending_id = 5;
  Uint128 user_data_128 = 6;
  uint64 user_data_64 = 7;
  uint32 user_data_32 = 8;
  uint32 timeout = 9;
  uint32 ledger = 10;
  uint32 code = 11;  // Note: uint16 in TigerBeetle, but proto3 uses uint32
  uint32 flags = 12; // Note: uint16 in TigerBeetle, but proto3 uses uint32
  uint64 timestamp = 13; // Nanoseconds since epoch
}

message TransferFlags {
  bool linked = 1;
  bool pending = 2;
  bool post_pending_transfer = 3;
  bool void_pending_transfer = 4;
  bool balancing_debit = 5;
  bool balancing_credit = 6;
  bool closing_debit = 7;
  bool closing_credit = 8;
  bool imported = 9;
}

// Get Account Transfers Request
message GetAccountTransfersRequest {
  Uint128 account_id = 1;
  uint32 code = 2;              // Optional filter by transfer code
  uint64 timestamp_min = 3;     // Inclusive, 0 = no lower bound
  uint64 timestamp_max = 4;     // Inclusive, 0 = no upper bound
  uint32 limit = 5;             // Max results (default: 8190)
  bool debits = 6;              // Include debits
  bool credits = 7;             // Include credits
  bool reversed = 8;            // Reverse chronological order
}

message GetAccountTransfersResponse {
  repeated Transfer transfers = 1;
  uint32 total_count = 2;
  bool has_more = 3;
  uint64 next_timestamp = 4;  // For pagination
}

// Query Transfers Request
message QueryTransfersRequest {
  Uint128 user_data_128 = 1;   // Optional, 0 = not filtered
  uint64 user_data_64 = 2;     // Optional, 0 = not filtered
  uint32 user_data_32 = 3;     // Optional, 0 = not filtered
  uint32 ledger = 4;           // Optional, 0 = not filtered
  uint32 code = 5;             // Optional, 0 = not filtered
  uint64 timestamp_min = 6;
  uint64 timestamp_max = 7;
  uint32 limit = 8;
  bool reversed = 9;
}

message QueryTransfersResponse {
  repeated Transfer transfers = 1;
  uint32 total_count = 2;
  bool has_more = 3;
}

// Lookup Transfers Request
message LookupTransfersRequest {
  repeated Uint128 ids = 1;
}

message LookupTransfersResponse {
  repeated Transfer transfers = 1;
}

// Filter for account-based queries
message AccountFilter {
  Uint128 account_id = 1;
  Uint128 user_data_128 = 2;
  uint64 user_data_64 = 3;
  uint32 user_data_32 = 4;
  uint32 code = 5;
  uint64 timestamp_min = 6;
  uint64 timestamp_max = 7;
  uint32 limit = 8;
  uint32 flags = 9;
}
```

## 5. Complete Go Example with HTTP API

```go
package main

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    tb "github.com/tigerbeetle/tigerbeetle-go"
    "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Server struct {
    client *tb.Client
}

// HTTP Handler for getting account transfers
func (s *Server) GetAccountTransfersHandler(w http.ResponseWriter, r *http.Request) {
    // Parse account ID from URL
    vars := mux.Vars(r)
    accountIDStr := vars["account_id"]
    
    accountID, err := types.HexStringToUint128(accountIDStr)
    if err != nil {
        http.Error(w, "Invalid account ID", http.StatusBadRequest)
        return
    }
    
    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    limit := uint32(100)
    if limitStr != "" {
        if l, err := strconv.ParseUint(limitStr, 10, 32); err == nil {
            limit = uint32(l)
        }
    }
    
    codeStr := r.URL.Query().Get("code")
    code := uint16(0)
    if codeStr != "" {
        if c, err := strconv.ParseUint(codeStr, 10, 16); err == nil {
            code = uint16(c)
        }
    }
    
    debits := r.URL.Query().Get("debits") == "true"
    credits := r.URL.Query().Get("credits") == "true"
    reversed := r.URL.Query().Get("reversed") == "true"
    
    // Default to showing both debits and credits
    if !debits && !credits {
        debits = true
        credits = true
    }
    
    // Create filter
    filterFlags := types.AccountFilterFlags{
        Debits:   debits,
        Credits:  credits,
        Reversed: reversed,
    }
    
    filter := types.AccountFilter{
        AccountID:    accountID,
        Code:         code,
        TimestampMin: 0,
        TimestampMax: 0,
        Limit:        limit,
        Flags:        filterFlags.ToUint32(),
    }
    
    // Query TigerBeetle
    transfers, err := s.client.GetAccountTransfers(filter)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Convert to response format
    response := TransferListResponse{
        Transfers:  make([]TransferResponse, len(transfers)),
        TotalCount: len(transfers),
        HasMore:    len(transfers) == int(limit),
    }
    
    for i, t := range transfers {
        response.Transfers[i] = ToTransferResponse(t)
    }
    
    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// HTTP Handler for querying transfers by code
func (s *Server) QueryTransfersByCodeHandler(w http.ResponseWriter, r *http.Request) {
    codeStr := r.URL.Query().Get("code")
    code, err := strconv.ParseUint(codeStr, 10, 16)
    if err != nil {
        http.Error(w, "Invalid code", http.StatusBadRequest)
        return
    }
    
    ledgerStr := r.URL.Query().Get("ledger")
    ledger, err := strconv.ParseUint(ledgerStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ledger", http.StatusBadRequest)
        return
    }
    
    filter := types.QueryFilter{
        Code:         uint16(code),
        Ledger:       uint32(ledger),
        TimestampMin: 0,
        TimestampMax: 0,
        Limit:        100,
        Flags:        types.QueryFilterFlags{Reversed: false}.ToUint32(),
    }
    
    transfers, err := s.client.QueryTransfers(filter)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := TransferListResponse{
        Transfers:  make([]TransferResponse, len(transfers)),
        TotalCount: len(transfers),
    }
    
    for i, t := range transfers {
        response.Transfers[i] = ToTransferResponse(t)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    // Initialize TigerBeetle client
    client, err := tb.NewClient(0, []string{"localhost:3000"}, 32)
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    server := &Server{client: client}
    
    // Setup routes
    r := mux.NewRouter()
    r.HandleFunc("/api/accounts/{account_id}/transfers", server.GetAccountTransfersHandler).Methods("GET")
    r.HandleFunc("/api/transfers", server.QueryTransfersByCodeHandler).Methods("GET")
    
    // Start server
    http.ListenAndServe(":8080", r)
}
```

## 6. Example API Usage

### Get transfers for an account
```bash
# Get all transfers for account
curl "http://localhost:8080/api/accounts/01936d89a0000000000000000000001/transfers"

# Get only debits
curl "http://localhost:8080/api/accounts/01936d89a0000000000000000000001/transfers?debits=true&credits=false"

# Get only credits, with specific code
curl "http://localhost:8080/api/accounts/01936d89a0000000000000000000001/transfers?credits=true&code=718"

# Get in reverse order (newest first)
curl "http://localhost:8080/api/accounts/01936d89a0000000000000000000001/transfers?reversed=true"

# Limit results
curl "http://localhost:8080/api/accounts/01936d89a0000000000000000000001/transfers?limit=50"
```

### Query transfers by code
```bash
curl "http://localhost:8080/api/transfers?code=718&ledger=1"
```

## 7. Summary

**Key Points:**

1. **get_account_transfers** - Best for querying all transfers for a specific account
   - Supports filtering by debits/credits
   - Can filter by transfer code
   - Supports time ranges
   - Results are paginated

2. **query_transfers** - Best for querying transfers by metadata
   - Filter by user_data fields
   - Filter by ledger and code
   - Time range support

3. **Data Types:**
   - ID fields (uint128) → `BYTEA` in PostgreSQL, `bytes` in Protobuf
   - CODE (uint16) → `SMALLINT` in PostgreSQL, `uint32` in Protobuf
   - All timestamps are uint64 nanoseconds since epoch

4. **Pagination:**
   - Use `limit` to control page size
   - Use `timestamp_min`/`timestamp_max` for time-based pagination
   - Check if more results exist by comparing result count to limit