package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// User represents a user entity
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/users", getUsers)
	e.GET("/users/:id", getUser)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	fmt.Println("Server started on port 8000")
	e.Start(":8000")
}

// DynamoDB client initialization
func getDynamoDBClient() *dynamodb.DynamoDB {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your desired AWS region
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}

	return dynamodb.New(sess)
}

// Retrieve all users
func getUsers(c echo.Context) error {
	svc := getDynamoDBClient()

	input := &dynamodb.ScanInput{
		TableName: aws.String("users"), // Replace with your DynamoDB table name
	}

	result, err := svc.Scan(input)
	if err != nil {
		log.Println("Failed to retrieve users:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve users")
	}

	var users []User
	for _, item := range result.Items {
		user := User{
			ID:   *item["id"].S,
			Name: *item["name"].S,
			Age:  int(*item["age"].N),
		}
		users = append(users, user)
	}

	return c.JSON(http.StatusOK, users)
}

// Retrieve a user by ID
func getUser(c echo.Context) error {
	svc := getDynamoDBClient()

	id := c.Param("id")

	input := &dynamodb.GetItemInput{
		TableName: aws.String("users"), // Replace with your DynamoDB table name
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		log.Println("Failed to retrieve user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user")
	}

	if result.Item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	user := User{
		ID:   *result.Item["id"].S,
		Name: *result.Item["name"].S,

		Age: int(*result.Item["age"].N),
	}

	return c.JSON(http.StatusOK, user)
}

// Create a new user
func createUser(c echo.Context) error {
	svc := getDynamoDBClient()

	user := new(User)
	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("users"), // Replace with your DynamoDB table name
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(user.ID),
			},
			"name": {
				S: aws.String(user.Name),
			},
			"age": {
				N: aws.String(fmt.Sprintf("%d", user.Age)),
			},
		},
	}

	_, err := svc.PutItem(input)
	if err != nil {
		log.Println("Failed to create user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(http.StatusCreated, user)
}

// Update an existing user
func updateUser(c echo.Context) error {
	svc := getDynamoDBClient()
	id := c.Param("id")

	user := new(User)
	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("users"), // Replace with your DynamoDB table name
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("set #n = :n, #a = :a"),
		ExpressionAttributeNames: map[string]*string{
			"#n": aws.String("name"),
			"#a": aws.String("age"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":n": {
				S: aws.String(user.Name),
			},
			":a": {
				N: aws.String(fmt.Sprintf("%d", user.Age)),
			},
		},
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		log.Println("Failed to update user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(http.StatusOK, user)
}

// Delete a user
func deleteUser(c echo.Context) error {
	svc := getDynamoDBClient()

	id := c.Param("id")

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("users"), // Replace with your DynamoDB table name
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		log.Println("Failed to delete user:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.NoContent(http.StatusNoContent)
}
