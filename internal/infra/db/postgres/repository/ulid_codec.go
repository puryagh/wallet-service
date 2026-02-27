package repository

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
)

// ULIDCodec is a custom codec for encoding/decoding ULID values with pgx v5.
// It implements both pgtype.Codec interfaces for proper bidirectional conversion.
type ULIDCodec struct{}

// FormatSupported returns true if the format is supported.
func (ULIDCodec) FormatSupported(format int16) bool {
	return format == pgtype.TextFormatCode || format == pgtype.BinaryFormatCode
}

// PreferredFormat returns the preferred format (text for ULID).
func (ULIDCodec) PreferredFormat() int16 {
	return pgtype.TextFormatCode
}

// PlanEncode creates a plan for encoding a value.
func (ULIDCodec) PlanEncode(m *pgtype.Map, oid uint32, format int16, value any) pgtype.EncodePlan {
	switch format {
	case pgtype.TextFormatCode:
		return &ulidEncodePlan{}
	case pgtype.BinaryFormatCode:
		return &ulidEncodePlan{}
	}
	return nil
}

// PlanScan creates a plan for scanning a value.
func (ULIDCodec) PlanScan(m *pgtype.Map, oid uint32, format int16, target any) pgtype.ScanPlan {
	switch format {
	case pgtype.TextFormatCode:
		return &ulidScanPlan{}
	case pgtype.BinaryFormatCode:
		return &ulidScanPlan{}
	}
	return nil
}

// DecodeDatabaseSQLValue decodes a database value to a driver.Value.
func (c ULIDCodec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return c.DecodeValue(m, oid, format, src)
}

// DecodeValue decodes a database value to any.
func (ULIDCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return nil, nil
	}

	var u ulid.ULID
	err := u.UnmarshalText(src)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ULID: %w", err)
	}

	return u, nil
}

// ulidEncodePlan is the encode plan for ULID values.
type ulidEncodePlan struct{}

// Encode encodes a ULID value to bytes.
func (ulidEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	var u ulid.ULID

	switch v := value.(type) {
	case ulid.ULID:
		u = v
	case *ulid.ULID:
		if v == nil {
			return nil, nil
		}
		u = *v
	default:
		return nil, fmt.Errorf("cannot encode %T to ULID", value)
	}

	// Convert ULID to string representation
	text, err := u.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ULID: %w", err)
	}

	buf = append(buf, text...)
	return buf, nil
}

// ulidScanPlan is the scan plan for ULID values.
type ulidScanPlan struct{}

// Scan scans a database value into a ULID.
func (ulidScanPlan) Scan(src []byte, dst any) error {
	if src == nil {
		return fmt.Errorf("cannot scan NULL into ULID")
	}

	switch d := dst.(type) {
	case *ulid.ULID:
		err := d.UnmarshalText(src)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ULID: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan into %T", dst)
	}
}

