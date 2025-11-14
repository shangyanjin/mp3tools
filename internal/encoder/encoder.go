package encoder

import (
	"fmt"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// DetectEncoding detects the character encoding of the given bytes
func DetectEncoding(data []byte) (string, error) {
	if len(data) == 0 {
		return "UTF-8", nil
	}

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return "", fmt.Errorf("failed to detect encoding: %w", err)
	}

	return result.Charset, nil
}

// ConvertToUTF8 converts bytes from the detected encoding to UTF-8
func ConvertToUTF8(data []byte, charset string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	// If already UTF-8, return as is
	if charset == "UTF-8" || charset == "" {
		return string(data), nil
	}

	// Get the decoder for the source charset
	decoder := getDecoder(charset)
	if decoder == nil {
		// If unknown charset, return as is
		return string(data), nil
	}

	// Decode to UTF-8
	decoded, err := decoder.Bytes(data)
	if err != nil {
		return string(data), fmt.Errorf("failed to decode from %s: %w", charset, err)
	}

	return string(decoded), nil
}

// ConvertStringToUTF8 detects encoding and converts string to UTF-8
func ConvertStringToUTF8(str string) (string, string, error) {
	data := []byte(str)

	// Detect encoding
	charset, err := DetectEncoding(data)
	if err != nil {
		return str, "UTF-8", err
	}

	// Convert to UTF-8
	utf8Str, err := ConvertToUTF8(data, charset)
	if err != nil {
		return str, charset, err
	}

	return utf8Str, charset, nil
}

// NeedsEncodingFix checks if the string needs encoding conversion
func NeedsEncodingFix(str string) bool {
	if str == "" {
		return false
	}

	data := []byte(str)
	charset, err := DetectEncoding(data)
	if err != nil {
		return false
	}

	return charset != "UTF-8" && charset != ""
}

// getDecoder returns the appropriate decoder for the given charset
func getDecoder(charset string) *encoding.Decoder {
	switch charset {
	case "GB2312", "GB-2312", "GBK", "GB18030":
		return simplifiedchinese.GBK.NewDecoder()
	case "Big5", "BIG5":
		return traditionalchinese.Big5.NewDecoder()
	case "UTF-16LE":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	case "UTF-16BE":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "ISO-8859-1", "windows-1252":
		// For ISO-8859-1, we can directly convert
		return nil
	default:
		return nil
	}
}

// FixEncoding is a convenience function that detects and fixes encoding
func FixEncoding(str string) (fixed string, originalCharset string, changed bool) {
	if str == "" {
		return "", "UTF-8", false
	}

	// First try to detect and fix double encoding (UTF-8 bytes misinterpreted as ISO-8859-1)
	if fixedStr, isDoubleEncoded := FixDoubleEncoding(str); isDoubleEncoded {
		return fixedStr, "UTF-8 (double-encoded)", true
	}

	// Then try normal encoding detection and conversion
	utf8Str, charset, err := ConvertStringToUTF8(str)
	if err != nil {
		return str, "UTF-8", false
	}

	changed = charset != "UTF-8" && utf8Str != str
	return utf8Str, charset, changed
}

// FixDoubleEncoding fixes double encoding issues where UTF-8 bytes were misinterpreted as ISO-8859-1
func FixDoubleEncoding(str string) (string, bool) {
	if str == "" {
		return str, false
	}

	// Convert string to bytes as if it were ISO-8859-1 (each rune becomes a byte)
	bytes := make([]byte, 0, len(str))
	for _, r := range str {
		if r > 255 {
			// If any rune is > 255, it's not double-encoded
			return str, false
		}
		bytes = append(bytes, byte(r))
	}

	// Try to interpret these bytes as UTF-8
	candidate := string(bytes)

	// Check if the result is valid UTF-8 and contains Chinese characters
	if !isValidUTF8WithChinese(candidate) {
		return str, false
	}

	return candidate, true
}

// isValidUTF8WithChinese checks if string is valid UTF-8 and contains Chinese characters
func isValidUTF8WithChinese(s string) bool {
	hasChinese := false
	for _, r := range s {
		// Check for Chinese characters (CJK Unified Ideographs)
		if r >= 0x4E00 && r <= 0x9FFF {
			hasChinese = true
		}
		// Also check for common Chinese punctuation
		if r >= 0x3000 && r <= 0x303F {
			hasChinese = true
		}
	}
	return hasChinese
}

// IsGarbled checks if a string appears to be garbled (unrecoverable)
func IsGarbled(str string) bool {
	if str == "" {
		return false
	}

	// Check if string contains many question marks or replacement characters
	questionMarkCount := 0
	replacementCharCount := 0
	invalidCharCount := 0

	for _, r := range str {
		if r == '?' {
			questionMarkCount++
		}
		if r == '\uFFFD' { // Unicode replacement character
			replacementCharCount++
		}
		if r > 0x10FFFF || (r >= 0xD800 && r <= 0xDFFF) {
			invalidCharCount++
		}
	}

	// If more than 30% of characters are question marks or replacement characters, likely garbled
	totalChars := len([]rune(str))
	if totalChars == 0 {
		return false
	}

	problemRatio := float64(questionMarkCount+replacementCharCount+invalidCharCount) / float64(totalChars)
	if problemRatio > 0.3 {
		return true
	}

	// Check if string contains many non-printable or unusual characters
	// that suggest encoding issues
	unusualCount := 0
	for _, r := range str {
		// Check for characters in problematic ranges that often indicate encoding issues
		if (r >= 0x80 && r <= 0x9F) || // Control characters in Latin-1
			(r >= 0x00 && r < 0x20 && r != '\n' && r != '\r' && r != '\t') {
			unusualCount++
		}
	}

	if float64(unusualCount)/float64(totalChars) > 0.2 {
		return true
	}

	return false
}