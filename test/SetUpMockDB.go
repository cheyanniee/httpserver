package test

import (
	"httpserver/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Setup Mock DB so as to not override the actual DB
func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	mockDB, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}

	// Override the global DB with mock
	models.DB = mockDB

	return mock
}
