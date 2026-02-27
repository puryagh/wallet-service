package repository

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

// TestULIDCodec_Encode tests the encoding of ULID values
func TestULIDCodec_Encode(t *testing.T) {
	codec := &ULIDCodec{}

	// Test ULID string: 01KJB94PH19MXWPQK9H030RKD2
	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	testULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse test ULID")

	// Test encoding
	plan := codec.PlanEncode(nil, 0, pgtype.TextFormatCode, testULID)
	require.NotNil(t, plan, "Encode plan should not be nil")

	buf := make([]byte, 0, 26)
	newBuf, err := plan.Encode(testULID, buf)
	require.NoError(t, err, "Failed to encode ULID")
	require.Equal(t, testULIDStr, string(newBuf), "Encoded ULID should match original string")
}

// TestULIDCodec_Scan tests the scanning of ULID values
func TestULIDCodec_Scan(t *testing.T) {
	codec := &ULIDCodec{}

	// Test ULID string: 01KJB94PH19MXWPQK9H030RKD2
	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	expectedULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse expected ULID")

	// Test scanning
	plan := codec.PlanScan(nil, 0, pgtype.TextFormatCode, &ulid.ULID{})
	require.NotNil(t, plan, "Scan plan should not be nil")

	var scannedULID ulid.ULID
	err = plan.Scan([]byte(testULIDStr), &scannedULID)
	require.NoError(t, err, "Failed to scan ULID")
	require.Equal(t, expectedULID, scannedULID, "Scanned ULID should match expected ULID")
}

// TestULIDCodec_RoundTrip tests encoding and then decoding
func TestULIDCodec_RoundTrip(t *testing.T) {
	codec := &ULIDCodec{}

	// Test ULID string: 01KJB94PH19MXWPQK9H030RKD2
	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	originalULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse original ULID")

	// Encode
	encodePlan := codec.PlanEncode(nil, 0, pgtype.TextFormatCode, originalULID)
	buf := make([]byte, 0, 26)
	encodedBuf, err := encodePlan.Encode(originalULID, buf)
	require.NoError(t, err, "Failed to encode ULID")

	// Decode
	scanPlan := codec.PlanScan(nil, 0, pgtype.TextFormatCode, &ulid.ULID{})
	var decodedULID ulid.ULID
	err = scanPlan.Scan(encodedBuf, &decodedULID)
	require.NoError(t, err, "Failed to decode ULID")

	// Verify round trip
	require.Equal(t, originalULID, decodedULID, "Round trip should preserve ULID value")
	require.Equal(t, testULIDStr, decodedULID.String(), "Round trip should preserve ULID string")
}

// TestULIDCodec_EncodePointer tests encoding of ULID pointer
func TestULIDCodec_EncodePointer(t *testing.T) {
	codec := &ULIDCodec{}

	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	testULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse test ULID")

	// Test encoding pointer
	plan := codec.PlanEncode(nil, 0, pgtype.TextFormatCode, &testULID)
	require.NotNil(t, plan, "Encode plan should not be nil")

	buf := make([]byte, 0, 26)
	newBuf, err := plan.Encode(&testULID, buf)
	require.NoError(t, err, "Failed to encode ULID pointer")
	require.Equal(t, testULIDStr, string(newBuf), "Encoded ULID pointer should match original string")
}

// TestULIDCodec_DecodeValue tests the DecodeValue method
func TestULIDCodec_DecodeValue(t *testing.T) {
	codec := &ULIDCodec{}

	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	expectedULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse expected ULID")

	// Test DecodeValue
	value, err := codec.DecodeValue(nil, 0, pgtype.TextFormatCode, []byte(testULIDStr))
	require.NoError(t, err, "Failed to decode value")
	require.IsType(t, ulid.ULID{}, value, "Decoded value should be ulid.ULID")

	decodedULID := value.(ulid.ULID)
	require.Equal(t, expectedULID, decodedULID, "Decoded value should match expected ULID")
}

// TestULIDCodec_NilPointer tests encoding of nil pointer
func TestULIDCodec_NilPointer(t *testing.T) {
	codec := &ULIDCodec{}

	plan := codec.PlanEncode(nil, 0, pgtype.TextFormatCode, (*ulid.ULID)(nil))
	require.NotNil(t, plan, "Encode plan should not be nil")

	buf := make([]byte, 0)
	newBuf, err := plan.Encode((*ulid.ULID)(nil), buf)
	require.NoError(t, err, "Failed to encode nil pointer")
	require.Nil(t, newBuf, "Encoded nil pointer should return nil")
}

// TestULIDCodec_Integration tests the codec with a mock pgx connection
func TestULIDCodec_Integration(t *testing.T) {
	// This test demonstrates how the codec would be registered
	// In actual usage, this would be done in main.go

	testULIDStr := "01KJB94PH19MXWPQK9H030RKD2"
	testULID, err := ulid.Parse(testULIDStr)
	require.NoError(t, err, "Failed to parse test ULID")

	// Create a type map (simulating pgx type registration)
	typeMap := pgtype.NewMap()
	typeMap.RegisterType(&pgtype.Type{
		Name:  "ulid",
		OID:   16702, // Custom ULID type OID
		Codec: &ULIDCodec{},
	})

	// Test that the codec is properly registered
	pgType, ok := typeMap.TypeForOID(16702)
	require.True(t, ok, "Type should be registered")
	require.NotNil(t, pgType, "Type should not be nil")
	require.NotNil(t, pgType.Codec, "Codec should be registered")

	// Test encoding through the type map
	encodePlan := pgType.Codec.PlanEncode(typeMap, 16702, pgtype.TextFormatCode, testULID)
	require.NotNil(t, encodePlan, "Encode plan should not be nil")

	buf := make([]byte, 0, 26)
	encodedBuf, err := encodePlan.Encode(testULID, buf)
	require.NoError(t, err, "Failed to encode through type map")
	require.Equal(t, testULIDStr, string(encodedBuf), "Encoded value should match")
}

// BenchmarkULIDCodec_Encode benchmarks the encoding performance
func BenchmarkULIDCodec_Encode(b *testing.B) {
	codec := &ULIDCodec{}
	testULID, _ := ulid.Parse("01KJB94PH19MXWPQK9H030RKD2")
	plan := codec.PlanEncode(nil, 0, pgtype.TextFormatCode, testULID)
	buf := make([]byte, 0, 26)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = plan.Encode(testULID, buf[:0])
	}
}

// BenchmarkULIDCodec_Scan benchmarks the scanning performance
func BenchmarkULIDCodec_Scan(b *testing.B) {
	codec := &ULIDCodec{}
	testULIDBytes := []byte("01KJB94PH19MXWPQK9H030RKD2")
	plan := codec.PlanScan(nil, 0, pgtype.TextFormatCode, &ulid.ULID{})
	var scannedULID ulid.ULID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = plan.Scan(testULIDBytes, &scannedULID)
	}
}
