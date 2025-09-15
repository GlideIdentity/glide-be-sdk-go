package glide

import (
	"regexp"
	"strings"
)

// ValidatePhoneNumber validates E.164 phone number format
// Returns an error if the phone number is invalid
func ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return nil // Phone number is optional for GetPhoneNumber
	}

	// E.164 format validation - strict, no cleaning
	if !strings.HasPrefix(phoneNumber, "+") {
		return NewError(ErrCodeInvalidPhoneNumber, "Phone number must be in E.164 format (start with +)")
	}

	if len(phoneNumber) < 8 {
		return NewError(ErrCodeInvalidPhoneNumber, "Phone number too short for E.164 format (minimum 8 characters including +)")
	}

	if len(phoneNumber) > 16 {
		return NewError(ErrCodeInvalidPhoneNumber, "Phone number too long for E.164 format (maximum 15 digits after +)")
	}

	// Check for any invalid characters (spaces, dashes, parentheses, etc.)
	// E.164 format only allows + followed by digits
	validFormat := regexp.MustCompile(`^\+\d+$`)
	if !validFormat.MatchString(phoneNumber) {
		return NewError(ErrCodeInvalidPhoneNumber, "Phone number contains invalid characters. E.164 format only allows + followed by digits")
	}

	// Detailed E.164 regex validation
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !e164Regex.MatchString(phoneNumber) {
		return NewError(ErrCodeInvalidPhoneNumber, "Invalid E.164 phone number format")
	}

	return nil
}

// ValidatePLMN validates PLMN (MCC/MNC) values
// Returns an error if the PLMN is invalid
func ValidatePLMN(plmn *PLMN) error {
	if plmn == nil {
		return nil // PLMN is optional
	}

	// MCC validation (3 digits) - no range check for telco labs
	mccRegex := regexp.MustCompile(`^\d{3}$`)
	if !mccRegex.MatchString(plmn.MCC) {
		return NewError(ErrCodeInvalidMCCMNC, "MCC must be exactly 3 digits")
	}

	// MNC validation (2 or 3 digits)
	mncRegex := regexp.MustCompile(`^\d{2,3}$`)
	if !mncRegex.MatchString(plmn.MNC) {
		return NewError(ErrCodeInvalidMCCMNC, "MNC must be 2 or 3 digits")
	}

	// No range validation - allowing unofficial MCCs for telco labs

	return nil
}

// ValidateUseCaseRequirements validates use case and phone number combination
// Returns an error if the requirements are not met
func ValidateUseCaseRequirements(useCase UseCase, phoneNumber string) error {
	// VerifyPhoneNumber requires a phone number
	if useCase == UseCaseVerifyPhoneNumber && phoneNumber == "" {
		return NewError(ErrCodeInvalidParameters, "Phone number is required for VerifyPhoneNumber use case")
	}

	return nil
}
