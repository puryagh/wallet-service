package card_iso

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	minPANLength               = 16
	maxPANLength               = 16
	defaultDiscretionaryLength = 7
	defaultServiceCode         = "201"
	defaultTrackNameMaxLength  = 26
)

// IssueSpec describes a magnetic-card issuance request.
type IssueSpec struct {
	IssuerBIN               string
	PANLength               int
	CVVLength               int
	ExpiryMonths            int
	ServiceCode             string
	CardholderName          string
	HMACSecret              string
	DiscretionaryDataLength int
	Random                  io.Reader
	Now                     time.Time
}

// IssuedCard holds the generated sensitive card material and its stored digests.
type IssuedCard struct {
	PAN            string
	MaskedPAN      string
	PANLast4       string
	ExpiryMonth    int
	ExpiryYear     int
	CardholderName string
	ServiceCode    string
	CVV            string
	CVV2           string
	Track1         string
	Track2         string
	PANFingerprint string
	CVVDigest      string
	CVV2Digest     string
	Track1Digest   string
	Track2Digest   string
}

// Issue generates a standard card PAN/CVV/track set and HMAC-backed storage digests.
func Issue(spec IssueSpec) (IssuedCard, error) {
	secret := strings.TrimSpace(spec.HMACSecret)
	if secret == "" {
		return IssuedCard{}, errors.New("card hmac secret is required")
	}

	rnd := spec.Random
	if rnd == nil {
		rnd = rand.Reader
	}

	now := spec.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	pan, err := GeneratePAN(spec.IssuerBIN, spec.PANLength, rnd)
	if err != nil {
		return IssuedCard{}, err
	}

	month, year, err := GenerateExpiry(now, spec.ExpiryMonths)
	if err != nil {
		return IssuedCard{}, err
	}

	cardholderName, err := NormalizeCardholderName(spec.CardholderName)
	if err != nil {
		return IssuedCard{}, err
	}

	serviceCode, err := NormalizeServiceCode(spec.ServiceCode)
	if err != nil {
		return IssuedCard{}, err
	}

	cvv, err := GenerateNumeric(spec.CVVLength, rnd)
	if err != nil {
		return IssuedCard{}, err
	}

	cvv2, err := GenerateNumeric(spec.CVVLength, rnd)
	if err != nil {
		return IssuedCard{}, err
	}

	discretionaryLength := spec.DiscretionaryDataLength
	if discretionaryLength == 0 {
		discretionaryLength = defaultDiscretionaryLength
	}

	discretionaryData, err := GenerateNumeric(discretionaryLength, rnd)
	if err != nil {
		return IssuedCard{}, err
	}

	track1, err := BuildTrack1(pan, cardholderName, month, year, serviceCode, discretionaryData)
	if err != nil {
		return IssuedCard{}, err
	}

	track2, err := BuildTrack2(pan, month, year, serviceCode, discretionaryData)
	if err != nil {
		return IssuedCard{}, err
	}

	monthText := fmt.Sprintf("%02d", month)
	yearText := fmt.Sprintf("%04d", year)

	panFingerprint, err := DigestStrings(secret, "pan_fingerprint", pan)
	if err != nil {
		return IssuedCard{}, err
	}

	cvvDigest, err := DigestStrings(secret, "cvv1", pan, monthText, yearText, serviceCode, cvv)
	if err != nil {
		return IssuedCard{}, err
	}

	cvv2Digest, err := DigestStrings(secret, "cvv2", pan, monthText, yearText, serviceCode, cvv2)
	if err != nil {
		return IssuedCard{}, err
	}

	track1Digest, err := DigestStrings(secret, "track1", track1)
	if err != nil {
		return IssuedCard{}, err
	}

	track2Digest, err := DigestStrings(secret, "track2", track2)
	if err != nil {
		return IssuedCard{}, err
	}

	maskedPAN, err := MaskPAN(pan)
	if err != nil {
		return IssuedCard{}, err
	}

	return IssuedCard{
		PAN:            pan,
		MaskedPAN:      maskedPAN,
		PANLast4:       pan[len(pan)-4:],
		ExpiryMonth:    month,
		ExpiryYear:     year,
		CardholderName: cardholderName,
		ServiceCode:    serviceCode,
		CVV:            cvv,
		CVV2:           cvv2,
		Track1:         track1,
		Track2:         track2,
		PANFingerprint: panFingerprint,
		CVVDigest:      cvvDigest,
		CVV2Digest:     cvv2Digest,
		Track1Digest:   track1Digest,
		Track2Digest:   track2Digest,
	}, nil
}

// NormalizePAN strips presentation separators and validates PAN characters.
func NormalizePAN(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("pan is required")
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || unicode.IsSpace(r):
			continue
		default:
			return "", fmt.Errorf("pan contains invalid character %q", r)
		}
	}

	normalized := b.String()
	if len(normalized) < minPANLength || len(normalized) > maxPANLength {
		return "", fmt.Errorf("pan length must be between %d and %d digits", minPANLength, maxPANLength)
	}

	return normalized, nil
}

// ValidatePAN validates PAN normalization and Luhn checksum.
func ValidatePAN(input string) error {
	pan, err := NormalizePAN(input)
	if err != nil {
		return err
	}

	if !isValidLuhn(pan) {
		return errors.New("pan failed luhn validation")
	}

	return nil
}

// LuhnCheckDigit computes the Luhn check digit for a PAN body.
func LuhnCheckDigit(body string) (int, error) {
	if body == "" {
		return 0, errors.New("pan body is required")
	}
	for _, r := range body {
		if r < '0' || r > '9' {
			return 0, errors.New("pan body must be numeric")
		}
	}

	sum := 0
	double := true
	for i := len(body) - 1; i >= 0; i-- {
		digit := int(body[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	return (10 - (sum % 10)) % 10, nil
}

// GeneratePAN generates a PAN for the given issuer BIN and total length.
func GeneratePAN(issuerBIN string, panLength int, rnd io.Reader) (string, error) {
	bin, err := normalizeDigitsOnly(issuerBIN, "issuer bin")
	if err != nil {
		return "", err
	}

	if len(bin) < 6 || len(bin) > 8 {
		return "", errors.New("issuer bin must be 6 to 8 digits")
	}
	if panLength < minPANLength || panLength > maxPANLength {
		return "", fmt.Errorf("pan length must be between %d and %d digits", minPANLength, maxPANLength)
	}
	if panLength <= len(bin)+1 {
		return "", errors.New("pan length is too small for issuer bin")
	}

	body, err := GenerateNumeric(panLength-len(bin)-1, rnd)
	if err != nil {
		return "", err
	}

	partial := bin + body
	checkDigit, err := LuhnCheckDigit(partial)
	if err != nil {
		return "", err
	}

	return partial + strconv.Itoa(checkDigit), nil
}

// GenerateNumeric creates a cryptographically random numeric string.
func GenerateNumeric(length int, rnd io.Reader) (string, error) {
	if length <= 0 {
		return "", errors.New("numeric length must be positive")
	}
	if rnd == nil {
		rnd = rand.Reader
	}

	out := make([]byte, 0, length)
	buf := make([]byte, length*2)
	for len(out) < length {
		if _, err := io.ReadFull(rnd, buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if b >= 250 {
				continue
			}
			out = append(out, '0'+(b%10))
			if len(out) == length {
				break
			}
		}
	}

	return string(out), nil
}

// GenerateExpiry returns a future expiry month/year.
func GenerateExpiry(now time.Time, months int) (int, int, error) {
	if months <= 0 {
		return 0, 0, errors.New("expiry months must be positive")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	expiry := now.UTC().AddDate(0, months, 0)
	return int(expiry.Month()), expiry.Year(), nil
}

// ValidateExpiry validates that a card expiry is structurally valid and not expired.
func ValidateExpiry(month, year int, now time.Time) error {
	if month < 1 || month > 12 {
		return errors.New("expiry month must be between 1 and 12")
	}
	if year < 2000 || year > 9999 {
		return errors.New("expiry year must be four digits")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	expiryEnd := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if now.UTC().After(expiryEnd) {
		return errors.New("card is expired")
	}

	return nil
}

// NormalizeServiceCode validates ISO service code formatting.
func NormalizeServiceCode(code string) (string, error) {
	serviceCode := strings.TrimSpace(code)
	if serviceCode == "" {
		serviceCode = defaultServiceCode
	}
	if len(serviceCode) != 3 {
		return "", errors.New("service code must be exactly 3 digits")
	}
	for _, r := range serviceCode {
		if r < '0' || r > '9' {
			return "", errors.New("service code must be numeric")
		}
	}
	return serviceCode, nil
}

// NormalizeCardholderName normalizes a cardholder name for storage and track generation.
func NormalizeCardholderName(name string) (string, error) {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if upper == "" {
		return "", errors.New("cardholder name is required")
	}

	var b strings.Builder
	lastWasSpace := false
	for _, r := range upper {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasSpace = false
		case unicode.IsSpace(r) || strings.ContainsRune("-./'", r):
			if !lastWasSpace && b.Len() > 0 {
				b.WriteByte(' ')
				lastWasSpace = true
			}
		default:
			return "", fmt.Errorf("cardholder name contains unsupported character %q", r)
		}
	}

	normalized := strings.TrimSpace(b.String())
	if normalized == "" {
		return "", errors.New("cardholder name is required")
	}
	runes := []rune(normalized)
	if len(runes) > defaultTrackNameMaxLength {
		normalized = string(runes[:defaultTrackNameMaxLength])
	}

	return normalized, nil
}

// MaskPAN masks the middle digits of a PAN, preserving the BIN and last 4.
func MaskPAN(input string) (string, error) {
	pan, err := NormalizePAN(input)
	if err != nil {
		return "", err
	}
	if len(pan) < 10 {
		return "", errors.New("pan is too short to mask")
	}
	return pan[:6] + strings.Repeat("*", len(pan)-10) + pan[len(pan)-4:], nil
}

// BuildTrack1 creates ISO/IEC 7813 Track 1 data.
func BuildTrack1(pan, cardholderName string, month, year int, serviceCode, discretionary string) (string, error) {
	if err := ValidatePAN(pan); err != nil {
		return "", err
	}
	if err := ValidateExpiry(month, year, time.Time{}); err != nil {
		return "", err
	}
	name, err := NormalizeCardholderName(cardholderName)
	if err != nil {
		return "", err
	}
	serviceCode, err = NormalizeServiceCode(serviceCode)
	if err != nil {
		return "", err
	}
	discretionary, err = normalizeOptionalDigits(discretionary, "track1 discretionary data")
	if err != nil {
		return "", err
	}

	track := fmt.Sprintf("%%B%s^%s^%02d%02d%s%s?", pan, name, year%100, month, serviceCode, discretionary)
	if len(track) > 79 {
		return "", errors.New("track1 exceeds maximum length")
	}
	return track, nil
}

// BuildTrack2 creates ISO/IEC 7813 Track 2 data.
func BuildTrack2(pan string, month, year int, serviceCode, discretionary string) (string, error) {
	if err := ValidatePAN(pan); err != nil {
		return "", err
	}
	if err := ValidateExpiry(month, year, time.Time{}); err != nil {
		return "", err
	}
	serviceCode, err := NormalizeServiceCode(serviceCode)
	if err != nil {
		return "", err
	}
	discretionary, err = normalizeOptionalDigits(discretionary, "track2 discretionary data")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(";%s=%02d%02d%s%s?", pan, year%100, month, serviceCode, discretionary), nil
}

// DigestStrings produces a hex HMAC digest for the provided namespace and values.
func DigestStrings(secret, namespace string, parts ...string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("digest secret is required")
	}
	if strings.TrimSpace(namespace) == "" {
		return "", errors.New("digest namespace is required")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(namespace))
	_, _ = mac.Write([]byte{0})
	for _, part := range parts {
		_, _ = mac.Write([]byte(part))
		_, _ = mac.Write([]byte{0})
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyDigest verifies a stored digest in constant time.
func VerifyDigest(secret, namespace, expected string, parts ...string) bool {
	actual, err := DigestStrings(secret, namespace, parts...)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func isValidLuhn(pan string) bool {
	sum := 0
	double := false
	for i := len(pan) - 1; i >= 0; i-- {
		digit := int(pan[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

func normalizeDigitsOnly(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("%s must be numeric", field)
		}
	}
	return trimmed, nil
}

func normalizeOptionalDigits(value, field string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeDigitsOnly(value, field)
}
