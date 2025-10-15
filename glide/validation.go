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
		return NewError(ErrCodeValidationError, "Phone number must be in E.164 format (start with +)")
	}

	if len(phoneNumber) < 8 {
		return NewError(ErrCodeValidationError, "Phone number too short for E.164 format (minimum 8 characters including +)")
	}

	if len(phoneNumber) > 16 {
		return NewError(ErrCodeValidationError, "Phone number too long for E.164 format (maximum 15 digits after +)")
	}

	// Check for any invalid characters (spaces, dashes, parentheses, etc.)
	// E.164 format only allows + followed by digits
	validFormat := regexp.MustCompile(`^\+\d+$`)
	if !validFormat.MatchString(phoneNumber) {
		return NewError(ErrCodeValidationError, "Phone number contains invalid characters. E.164 format only allows + followed by digits")
	}

	// Detailed E.164 regex validation
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !e164Regex.MatchString(phoneNumber) {
		return NewError(ErrCodeValidationError, "Invalid E.164 phone number format")
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
		return NewError(ErrCodeValidationError, "MCC must be exactly 3 digits")
	}

	// MNC validation (2 or 3 digits)
	mncRegex := regexp.MustCompile(`^\d{2,3}$`)
	if !mncRegex.MatchString(plmn.MNC) {
		return NewError(ErrCodeValidationError, "MNC must be 2 or 3 digits")
	}

	// No range validation - allowing unofficial MCCs for telco labs

	return nil
}

// ValidateConsentData validates consent data if provided
func ValidateConsentData(consent *ConsentData) error {
	if consent == nil {
		return nil
	}

	// All fields are required if consent data is provided
	if consent.ConsentText == "" {
		return NewError(ErrCodeValidationError, "Consent text is required")
	}
	if consent.PolicyLink == "" {
		return NewError(ErrCodeValidationError, "Policy link is required")
	}
	if consent.PolicyText == "" {
		return NewError(ErrCodeValidationError, "Policy text is required")
	}

	// Validate policy link is a valid URL
	if !strings.HasPrefix(consent.PolicyLink, "http://") && !strings.HasPrefix(consent.PolicyLink, "https://") {
		return NewError(ErrCodeValidationError, "Policy link must be a valid URL")
	}

	return nil
}

// ValidateUseCaseRequirements validates use case and phone number/PLMN combination
// Returns an error if the requirements are not met
func ValidateUseCaseRequirements(useCase UseCase, phoneNumber string, plmn *PLMN) error {
	switch useCase {
	case UseCaseGetPhoneNumber:
		// GetPhoneNumber: We don't know the phone number, need PLMN
		if phoneNumber != "" {
			return NewError(ErrCodeValidationError, "Phone number should not be provided for GetPhoneNumber use case")
		}
		if plmn == nil {
			return NewError(ErrCodeMissingParameters, "PLMN (MCC/MNC) is required for GetPhoneNumber use case")
		}

	case UseCaseVerifyPhoneNumber:
		// VerifyPhoneNumber: Need phone number, PLMN is optional
		if phoneNumber == "" {
			return NewError(ErrCodeMissingParameters, "Phone number is required for VerifyPhoneNumber use case")
		}
	}

	return nil
}
