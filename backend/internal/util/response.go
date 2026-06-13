package util

import "github.com/gin-gonic/gin"

// ErrorBody is the standard error envelope returned by the API.
type ErrorBody struct {
	Error string `json:"error"`
}

// OK writes a 200 response with the given payload.
func OK(c *gin.Context, payload any) {
	c.JSON(200, payload)
}

// Created writes a 201 response with the given payload.
func Created(c *gin.Context, payload any) {
	c.JSON(201, payload)
}

// Fail writes an error response with the given status code and message.
func Fail(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorBody{Error: msg})
}
