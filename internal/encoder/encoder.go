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
	unusualCharCount := 0
	latin1ExtendedCount := 0 // Latin-1 extended characters (0xA0-0xFF) that shouldn't appear in Chinese text

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
		// Check for unusual characters that often indicate encoding issues
		// Latin-1 control characters, multiplication sign (×), division sign (÷), etc.
		if (r >= 0x80 && r <= 0x9F) || // Control characters in Latin-1
			r == 0xD7 || r == 0xF7 || // × and ÷
			(r >= 0x00 && r < 0x20 && r != '\n' && r != '\r' && r != '\t') {
			unusualCharCount++
		}
		// Check for Latin-1 extended characters (0xA0-0xFF) that often indicate encoding issues
		// These characters shouldn't appear in Chinese text
		// If we see many of these, it's likely a GBK/GB2312 string misinterpreted as Latin-1
		if r >= 0xA0 && r <= 0xFF && r != 0xA0 { // Exclude non-breaking space
			latin1ExtendedCount++
		}
	}

	totalChars := len([]rune(str))
	if totalChars == 0 {
		return false
	}

	// If more than 10% of characters are question marks, likely garbled
	questionRatio := float64(questionMarkCount) / float64(totalChars)
	if questionRatio > 0.1 {
		return true
	}

	// If more than 20% of characters are problematic, likely garbled
	problemRatio := float64(questionMarkCount+replacementCharCount+invalidCharCount+unusualCharCount+latin1ExtendedCount) / float64(totalChars)
	if problemRatio > 0.2 {
		return true
	}

	// Check for patterns that indicate garbled text
	// If contains many unusual characters (like ×, ÷, etc.) mixed with question marks
	if questionMarkCount > 0 && (unusualCharCount > 0 || latin1ExtendedCount > 0) {
		// If both question marks and unusual chars exist, likely garbled
		if float64(questionMarkCount+unusualCharCount+latin1ExtendedCount)/float64(totalChars) > 0.2 {
			return true
		}
	}

	// If contains many Latin-1 extended characters (likely encoding issue)
	if latin1ExtendedCount > 0 {
		latin1Ratio := float64(latin1ExtendedCount) / float64(totalChars)
		if latin1Ratio > 0.3 {
			return true
		}
	}

	return false
}

// isValidLatin1Char checks if a character is a valid Latin-1 character that might legitimately appear
func isValidLatin1Char(r rune) bool {
	// Common Latin-1 characters that might appear in mixed text
	// This is a conservative list - if in doubt, consider it garbled
	validChars := []rune{
		0xC0, 0xC1, 0xC2, 0xC3, 0xC4, 0xC5, // À Á Â Ã Ä Å
		0xC8, 0xC9, 0xCA, 0xCB,             // È É Ê Ë
		0xCC, 0xCD, 0xCE, 0xCF,             // Ì Í Î Ï
		0xD0, 0xD1,                         // Ð Ñ
		0xD2, 0xD3, 0xD4, 0xD5, 0xD6,       // Ò Ó Ô Õ Ö
		0xD9, 0xDA, 0xDB, 0xDC,             // Ù Ú Û Ü
		0xDD, 0xDE,                         // Ý Þ
		0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, // à á â ã ä å
		0xE8, 0xE9, 0xEA, 0xEB,             // è é ê ë
		0xEC, 0xED, 0xEE, 0xEF,             // ì í î ï
		0xF0, 0xF1,                         // ð ñ
		0xF2, 0xF3, 0xF4, 0xF5, 0xF6,       // ò ó ô õ ö
		0xF9, 0xFA, 0xFB, 0xFC,             // ù ú û ü
		0xFD, 0xFE, 0xFF,                   // ý þ ÿ
	}
	for _, valid := range validChars {
		if r == valid {
			return true
		}
	}
	return false
}