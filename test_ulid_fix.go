// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/oklog/ulid/v2"
)

func main() {
	// Get database connection string from environment or use default
	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		dbSource = "postgresql://postgres:postgres@localhost:5432/wallet_service?sslmode=disable"
	}

	fmt.Println("🔧 Testing ULID Fix with Database Connection")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Printf("Database: %s\n\n", dbSource)

	// Parse the connection string
	config, err := pgx.ParseConfig(dbSource)
	if err != nil {
		log.Fatalf("❌ Failed to parse database config: %v", err)
	}

	// Register ULID codec
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "ulid",
			OID:   16702,
			Codec: &repository.ULIDCodec{},
		})
		return nil
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("✅ Connected to database successfully")
	fmt.Println()

	// Test with the existing user ULID
	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	fmt.Printf("🧪 Testing with existing user ULID: %s\n", testULIDStr)
	fmt.Println()

	// Parse the ULID
	userIdentifier, err := ulid.Parse(testULIDStr)
	if err != nil {
		log.Fatalf("❌ Failed to parse ULID: %v", err)
	}

	fmt.Printf("📝 Parsed ULID: %s\n", userIdentifier.String())
	fmt.Printf("   Binary representation: %x\n", userIdentifier[:])
	fmt.Println()

	// Test 1: Query by user_identifier
	fmt.Println("Test 1: Query account by user_identifier")
	fmt.Println("-" + string(make([]byte, 40)))

	var account struct {
		ID                  int64
		Identifier          ulid.ULID
		Title               string
		UserIdentifier      ulid.ULID
		BaseAssetIdentifier ulid.ULID
		Banned              bool
	}

	query := `SELECT id, identifier, title, user_identifier, base_asset_identifier, banned 
	          FROM accounts 
	          WHERE user_identifier = $1 
	          LIMIT 1`

	err = conn.QueryRow(ctx, query, userIdentifier).Scan(
		&account.ID,
		&account.Identifier,
		&account.Title,
		&account.UserIdentifier,
		&account.BaseAssetIdentifier,
		&account.Banned,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("⚠️  No account found for this user_identifier")
		} else {
			log.Fatalf("❌ Query failed: %v", err)
		}
	} else {
		fmt.Println("✅ Query successful!")
		fmt.Printf("   Account ID: %d\n", account.ID)
		fmt.Printf("   Account Identifier: %s\n", account.Identifier.String())
		fmt.Printf("   Title: %s\n", account.Title)
		fmt.Printf("   User Identifier: %s\n", account.UserIdentifier.String())
		fmt.Printf("   Base Asset Identifier: %s\n", account.BaseAssetIdentifier.String())
		fmt.Printf("   Banned: %v\n", account.Banned)
	}
	fmt.Println()

	// Test 2: Query by identifier (the original failing query)
	fmt.Println("Test 2: Query account by identifier (GetUserAccount)")
	fmt.Println("-" + string(make([]byte, 40)))

	// First, get an account identifier to test with
	var testIdentifier ulid.ULID
	err = conn.QueryRow(ctx, "SELECT identifier FROM accounts LIMIT 1").Scan(&testIdentifier)
	if err != nil {
		log.Fatalf("❌ Failed to get test identifier: %v", err)
	}

	fmt.Printf("   Testing with identifier: %s\n", testIdentifier.String())

	query2 := `SELECT id, identifier, title, user_identifier, base_asset_identifier, banned 
	           FROM accounts 
	           WHERE identifier = $1 
	           LIMIT 1`

	var account2 struct {
		ID                  int64
		Identifier          ulid.ULID
		Title               string
		UserIdentifier      ulid.ULID
		BaseAssetIdentifier ulid.ULID
		Banned              bool
	}

	err = conn.QueryRow(ctx, query2, testIdentifier).Scan(
		&account2.ID,
		&account2.Identifier,
		&account2.Title,
		&account2.UserIdentifier,
		&account2.BaseAssetIdentifier,
		&account2.Banned,
	)

	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}

	fmt.Println("✅ Query successful!")
	fmt.Printf("   Account ID: %d\n", account2.ID)
	fmt.Printf("   Account Identifier: %s\n", account2.Identifier.String())
	fmt.Printf("   Title: %s\n", account2.Title)
	fmt.Printf("   User Identifier: %s\n", account2.UserIdentifier.String())
	fmt.Println()

	// Summary
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println("🎉 All tests passed! ULID fix is working correctly!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("  ✅ ULID encoding (Go → PostgreSQL) works")
	fmt.Println("  ✅ ULID decoding (PostgreSQL → Go) works")
	fmt.Println("  ✅ Query with ULID parameter works")
	fmt.Println("  ✅ GetUserAccount query pattern works")
	fmt.Println()
	fmt.Println("The fix successfully resolves the original error:")
	fmt.Println("  'invalid input syntax for type ulid: \"\\x019c96...\"'")
}

