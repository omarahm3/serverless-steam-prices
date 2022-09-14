package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/omarahm3/serverless-steam-prices/pkg/handlers"
)

type (
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler, err := handlers.Prepare(req)
	if err != nil {
		return handlers.JSONResponse(http.StatusInternalServerError, err.Error())
	}

	return handler.GetAppDetailsOnTheFly()
}

func main() {
	lambda.Start(handler)
}
