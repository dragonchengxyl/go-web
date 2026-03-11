package crypto

import (
	"fmt"
	"regexp"
	"unicode"
)

// PasswordStrength represents password strength requirements
type PasswordStrength struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordStrength returns default password strength requirements
func DefaultPasswordStrength() PasswordStrength {
	return PasswordStrength{
		MinLength:      8,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false,
	}
}

// ValidatePassword validates password against strength requirements
func ValidatePassword(password string, strength PasswordStrength) error {
	if len(password) < strength.MinLength {
		return fmt.Errorf("密码长度至少需要 %d 个字符", strength.MinLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if strength.RequireUpper && !hasUpper {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}

	if strength.RequireLower && !hasLower {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}

	if strength.RequireNumber && !hasNumber {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	if strength.RequireSpecial && !hasSpecial {
		return fmt.Errorf("密码必须包含至少一个特殊字符")
	}

	return nil
}

// IsCommonPassword checks if password is in common password list
// In production, this should check against a comprehensive list
func IsCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "111111", "iloveyou", "master", "sunshine",
		"ashley", "bailey", "passw0rd", "shadow", "123123",
		"654321", "superman", "qazwsx", "michael", "football",
	}

	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}

	return false
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}
	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("用户名长度至少需要 3 个字符")
	}

	if len(username) > 20 {
		return fmt.Errorf("用户名长度不能超过 20 个字符")
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("用户名只能包含字母、数字、下划线和连字符")
	}

	return nil
}
