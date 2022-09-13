package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
)

func JSONResponse(status int, body interface{}) (events.APIGatewayProxyResponse, error) {
	strBody, _ := json.Marshal(body)

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(strBody),
	}, nil
}

func errorResponse(err error) (Response, error) {
	return JSONResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
}
