package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/omarahm3/serverless-steam-prices/pkg/handlers"
)

type (
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler := &handlers.Handler{
		Req: req,
	}

	return handler.GetAppDetailsOnTheFly()
}

func main() {
	lambda.Start(handler)
}
