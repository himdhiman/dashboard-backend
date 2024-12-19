package helpers

import "regexp"

func IsValidDatabaseName(name string) bool {
	// MongoDB database names cannot be empty and cannot contain '/', '\', '.', '"', '*', '<', '>', ':', '|', '?'
	if name == "" {
		return false
	}
	invalidChars := regexp.MustCompile(`[\/\\\.\*"<>\:\|\?]`)
	return !invalidChars.MatchString(name)
}

// isValidCollectionName validates the collection name
func IsValidCollectionName(name string) bool {
	// MongoDB collection names cannot be empty and cannot contain '$'
	if name == "" {
		return false
	}
	invalidChars := regexp.MustCompile(`[$]`)
	return !invalidChars.MatchString(name)
}
