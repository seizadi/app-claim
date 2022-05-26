package graphdb

import (
	"encoding/json"
)

type StatusCode int

const (
	NotFound StatusCode = iota
	AlreadyExists
)


type DomainError struct {
	statusCode StatusCode
	message    string
	details    map[string]interface{}
}

func NewDomainError(statusCode StatusCode, message string, details map[string]interface{}) error {
	return &DomainError{
		statusCode: statusCode,
		message:    message,
		details:    details,
	}
}

func (d *DomainError) Error() string {
	errorJson, _ := json.Marshal(map[string]interface{}{
		"status":  "error",
		"code":    d.statusCode,
		"message": d.message,
		"details": d.details,
	})
	return string(errorJson)
}

func (d *DomainError) StatusCode() StatusCode {
	return d.statusCode
}

