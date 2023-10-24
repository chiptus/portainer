package liblicense

// Errors are return from license generation API
const (
	LicenseExistsError            = "License already exists"
	LicenseGenerationError        = "Failed generating license"
	LicenseDatabaseError          = "Failed saving license to database"
	LicensePayloadValidationError = "Invalid license payload"
)
