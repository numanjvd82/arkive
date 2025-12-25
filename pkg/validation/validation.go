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

func PasswordError(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters."
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
		return "Password must include a lowercase letter."
	}
	if !hasUpper {
		return "Password must include an uppercase letter."
	}
	if !hasSymbol {
		return "Password must include a symbol."
	}

	return ""
}
