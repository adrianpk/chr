package model

// ValidateParsedResponse keeps parsing and semantic failures on the shared response path.
func ValidateParsedResponse(response Response, parseErr error) error {
	if parseErr != nil {
		return parseErr
	}

	return ValidateResponse(response)
}
