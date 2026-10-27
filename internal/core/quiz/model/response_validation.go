package model

func ValidateParsedResponse(response QuizModelResponse, parseErr error) error {
	if parseErr != nil {
		return parseErr
	}

	return ValidateQuizModelResponse(response)
}
