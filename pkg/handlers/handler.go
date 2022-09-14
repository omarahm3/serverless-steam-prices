package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/omarahm3/serverless-steam-prices/pkg/app"
)

const (
	MINIMUM_QUERY_CHARS = 2
)

type (
	Json     map[string]interface{}
	Request  = events.APIGatewayProxyRequest
	Response = events.APIGatewayProxyResponse
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

type Handler struct {
	Req       Request
	TableName string
	Client    *dynamodb.DynamoDB
}

func Prepare(req Request) (*Handler, error) {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic(fmt.Errorf("TABLE_NAME env variable is not set"))
	}

	dynamoClient, err := getDynamoClient()
	if err != nil {
		return nil, fmt.Errorf("could not establish db connection")
	}

	return newHandler(req, tableName, dynamoClient), nil
}

func (h *Handler) GetAppDetails() (Response, error) {
	// Get the path parameter that was sent
	query := h.Req.QueryStringParameters["query"]

	if query == "" {
		return JSONResponse(http.StatusBadGateway, Json{"message": "you must supply 'query' parameter"})
	}

	if len(query) <= MINIMUM_QUERY_CHARS {
		return JSONResponse(http.StatusNotFound, Json{"message": "query must be more than 2 characters"})
	}

	apps, err := app.GetAllAppsByName(query, h.TableName, h.Client)
	if err != nil {
		return JSONResponse(http.StatusBadGateway, Json{"message": fmt.Sprintf("error getting apps with this query: [%s]", query)})
	}

	if len(apps) == 0 {
		return JSONResponse(http.StatusOK, Json{"message": fmt.Sprintf("no games were found using this query: [%s]", query)})
	}

	games, err := app.GetAllAppsDetails(apps)
	if err != nil {
		return JSONResponse(http.StatusBadGateway, Json{"message": "error occurred while getting app details"})
	}

	return JSONResponse(http.StatusOK, Json{
		"total": len(games),
		"apps":  games,
	})
}

func (h *Handler) GetAppDetailsOnTheFly() (Response, error) {
	// Get the path parameter that was sent
	query := h.Req.QueryStringParameters["query"]

	if query == "" {
		return JSONResponse(http.StatusBadGateway, Json{"message": "you must supply 'query' parameter"})
	}

	if len(query) <= MINIMUM_QUERY_CHARS {
		return JSONResponse(http.StatusNotFound, Json{"message": "query must be more than 2 characters"})
	}

	apps, err := app.GetAllGames()
	if err != nil {
		return JSONResponse(http.StatusBadGateway, Json{"message": fmt.Sprintf("error getting apps with this query: [%s]", query)})
	}

	apps = app.Format(apps)
	found, err := app.LookFor(query, apps)
	if err != nil {
		return JSONResponse(http.StatusBadGateway, Json{"message": fmt.Sprintf("error filtering apps with this query: [%s]", query)})
	}

	if len(found) == 0 {
		return JSONResponse(http.StatusOK, Json{"message": fmt.Sprintf("no games were found using this query: [%s]", query)})
	}

	return JSONResponse(http.StatusOK, Json{
		"total": len(found),
		"apps":  found,
	})
}

func getDynamoClient() (*dynamodb.DynamoDB, error) {
	var config *aws.Config
	region := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")

	// In case function is running locally
	if awsAccessKey == "" {
		config = &aws.Config{
			Region:   aws.String(region),
			Endpoint: aws.String("http://localhost:8001"),
		}
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return dynamodb.New(awsSession, config), nil
}

func newHandler(req Request, tableName string, dynamoClient *dynamodb.DynamoDB) *Handler {
	return &Handler{
		Req:       req,
		TableName: tableName,
		Client:    dynamoClient,
	}
}
