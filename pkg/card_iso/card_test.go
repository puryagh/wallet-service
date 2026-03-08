package card_iso

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGeneratePANAndValidate(t *testing.T) {
	pan, err := GeneratePAN("502229", 16, bytes.NewReader(make([]byte, 64)))
	require.NoError(t, err)
	require.Len(t, pan, 16)
	require.True(t, strings.HasPrefix(pan, "502229"))
	require.NoError(t, ValidatePAN(pan))
}

func TestIssueGeneratesSecureCardMaterial(t *testing.T) {
	now := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)
	issued, err := Issue(IssueSpec{
		IssuerBIN:               "502229",
		PANLength:               16,
		CVVLength:               3,
		ExpiryMonths:            36,
		ServiceCode:             "201",
		CardholderName:          "Jane Doe",
		HMACSecret:              "top-secret",
		DiscretionaryDataLength: 7,
		Random:                  bytes.NewReader(make([]byte, 128)),
		Now:                     now,
	})
	require.NoError(t, err)

	require.NoError(t, ValidatePAN(issued.PAN))
	require.Equal(t, "JANE DOE", issued.CardholderName)
	require.Equal(t, "201", issued.ServiceCode)
	require.Equal(t, 3, len(issued.CVV))
	require.Equal(t, 3, len(issued.CVV2))
	require.Equal(t, issued.PAN[len(issued.PAN)-4:], issued.PANLast4)
	require.Equal(t, issued.MaskedPAN, "502229******0000")
	require.Equal(t, 1, issued.ExpiryMonth)
	require.Equal(t, 2029, issued.ExpiryYear)
	require.Contains(t, issued.Track1, "%B")
	require.Contains(t, issued.Track2, ";")
	require.NotEmpty(t, issued.PANFingerprint)
	require.True(t, VerifyDigest("top-secret", "pan_fingerprint", issued.PANFingerprint, issued.PAN))
	require.True(t, VerifyDigest("top-secret", "track2", issued.Track2Digest, issued.Track2))
	monthText := "01"
	yearText := "2029"
	require.True(t, VerifyDigest("top-secret", "cvv1", issued.CVVDigest, issued.PAN, monthText, yearText, issued.ServiceCode, issued.CVV))
	require.True(t, VerifyDigest("top-secret", "cvv2", issued.CVV2Digest, issued.PAN, monthText, yearText, issued.ServiceCode, issued.CVV2))
}

func TestNormalizeCardholderName(t *testing.T) {
	name, err := NormalizeCardholderName("  jane-doe/iii  ")
	require.NoError(t, err)
	require.Equal(t, "JANE DOE III", name)
}

func TestNormalizeCardholderNameSupportsPersian(t *testing.T) {
	name, err := NormalizeCardholderName("  علی-رضا/مرادی  ")
	require.NoError(t, err)
	require.Equal(t, "علی رضا مرادی", name)
}

func TestValidateExpiry(t *testing.T) {
	now := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, ValidateExpiry(6, 2026, now))
	require.Error(t, ValidateExpiry(5, 2026, now))
	require.Error(t, ValidateExpiry(13, 2026, now))
}

func TestBuildTrackDataRejectsInvalidInputs(t *testing.T) {
	_, err := BuildTrack1("1234", "JANE DOE", 1, 2028, "20A", "123")
	require.Error(t, err)

	_, err = BuildTrack2("1234", 1, 2028, "201", "ABC")
	require.Error(t, err)
}

func TestVerifyDigestRejectsMismatchedData(t *testing.T) {
	digest, err := DigestStrings("top-secret", "sample", "one")
	require.NoError(t, err)
	require.True(t, VerifyDigest("top-secret", "sample", digest, "one"))
	require.False(t, VerifyDigest("top-secret", "sample", digest, "two"))
}
