package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestStartSudoku(t *testing.T) {
	handler := StartSudokuHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)
	userId := os.Getenv("TEST_USER_ID")

	// happy path
	obj := e.POST("/").WithHeader("Authorization", userId).Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("Id", "StartState")
}
