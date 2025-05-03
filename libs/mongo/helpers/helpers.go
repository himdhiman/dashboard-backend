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


func SetUpdate(fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"$set": fields}
}

func PullUpdate(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{"$pull": map[string]interface{}{field: value}}
}

func PushUpdate(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{"$push": map[string]interface{}{field: value}}
}