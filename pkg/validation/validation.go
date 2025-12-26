package validation

import "unicode"

type Errors map[string]string

const GeneralKey = "general"

func New() Errors {
	return Errors{}
}

func (errs Errors) Add(key, message string) {
	if errs == nil {
		return
	}
	if key == "" || message == "" {
		return
	}
	errs[key] = message
}

func (errs Errors) HasAny() bool {
	return len(errs) > 0
}

func FieldError(errors Errors, key string) string {
	if errors == nil {
		return ""
	}
	return errors[key]
}

type PasswordIssue string

const (
	PasswordTooShort      PasswordIssue = "too_short"
	PasswordMissingLower  PasswordIssue = "missing_lower"
	PasswordMissingUpper  PasswordIssue = "missing_upper"
	PasswordMissingSymbol PasswordIssue = "missing_symbol"
)

func PasswordIssueFor(password string) PasswordIssue {
	if len(password) < 8 {
		return PasswordTooShort
	}

	hasLower := false
	hasUpper := false
	hasSymbol := false

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if !hasLower {
		return PasswordMissingLower
	}
	if !hasUpper {
		return PasswordMissingUpper
	}
	if !hasSymbol {
		return PasswordMissingSymbol
	}

	return ""
}
