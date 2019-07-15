package api

import (
	"fmt"
	"net/http"
	"taskboard-api-go/service"

	"github.com/gin-gonic/gin"
)

// HandleErrorStatus sets http status and error response corresponding service.SvcError
func SetErrorStatus(c *gin.Context, err error) {
	serr, ok := err.(*service.SvcError)
	if !ok {
		serr = service.NewSvcErrorf(service.ErrorCodeUnexpected, err,
			"Unexpected error occurred Error:%s", err.Error()).(*service.SvcError)
	}
	// logging to stdout
	if serr.Cause == nil {
		fmt.Printf("Service error occurred. Code:%s Message:%s", serr.Code, serr.Message)
	} else {
		fmt.Printf("Service error occurred. Code:%s Message:%s Cause:%+v", serr.Code, serr.Message, serr.Cause)
	}
	var status int
	switch serr.Code {
	case service.ErrorCodeUnexpected:
		status = http.StatusInternalServerError
	case service.ErrorCodeBadRequest:
		status = http.StatusBadRequest
	case service.ErrorCodeInvalidArguments:
		status = http.StatusNotAcceptable
	case service.ErrorCodeInvalidlStatus:
		status = http.StatusInternalServerError
	case service.ErrorCodeDB:
		status = http.StatusInternalServerError
	case service.ErrorCodeNotFound:
		status = http.StatusNotFound
	case service.ErrorCodeAlreadyExist:
		status = http.StatusConflict
	case service.ErrorCodeOptimisticLockFailure:
		status = http.StatusPreconditionFailed
	case service.ErrorCodePreconditionInvalid:
		status = http.StatusPreconditionFailed
	case service.ErrorCodeUnauthenticated:
		status = http.StatusUnauthorized
	}
	errorResponse := &ErrorResponse{
		Code:    string(serr.Code),
		Message: serr.Message,
		Details: serr.Details,
	}
	c.IndentedJSON(status, errorResponse)
}
