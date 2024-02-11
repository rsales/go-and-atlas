package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

// Define your MongoDB connection string
var uri = "mongodb+srv://" + goDotEnvVariable("USERNAME") + ":" + goDotEnvVariable("PASSWORD") + "@" + goDotEnvVariable("CLUSTER") + "/?retryWrites=true&w=majorit"

// Create a global variable to hold our MongoDB connection
var mongoClient *mongo.Client

// This function runs before we call our main function and connects to our MongoDB database. If it cannot connect, the application stops.
func init() {
	if err := connectToMongoDB(); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}
}

// Our entry point into our application
func main() {
	// The simplest way to start a Gin application using the frameworks defaults
	route := gin.Default()

	// Our route definitions
	route.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})
	route.GET("/movies", getMovies)
	route.GET("/movies/:id", getMovieByID)
	route.POST("/movies/aggregations", aggregateMovies)

	// The Run() method starts our Gin server
	route.Run()
}

// Implemention of the /movies route that returns all of the movies from our movies collection.
func getMovies(c *gin.Context) {
	// Find movies
	cursor, err := mongoClient.Database("sample_mflix").Collection("movies").Find(context.TODO(), bson.D{{}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map results
	var movies []bson.M
	if err = cursor.All(context.TODO(), &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return movies
	c.JSON(http.StatusOK, movies)
}

// The implementation of our /movies/{id} endpoint that returns a single movie based on the provided ID
func getMovieByID(c *gin.Context) {
	// Get movie ID from URL
	idStr := c.Param("id")

	// Convert id string to ObjectId
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find movie by ObjectId
	var movie bson.M
	err = mongoClient.Database("sample_mflix").Collection("movies").FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return movie
	c.JSON(http.StatusOK, movie)
}

// The implementation of our /movies/aggregations endpoint that allows a user to pass in an aggregation to run our the movies collection.
func aggregateMovies(c *gin.Context) {
	// Get aggregation pipeline from request body
	var pipeline interface{}
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Run aggregations
	cursor, err := mongoClient.Database("sample_mflix").Collection("movies").Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map results
	var result []bson.M
	if err = cursor.All(context.TODO(), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return result
	c.JSON(http.StatusOK, result)
}

func connectToMongoDB() error {
	serverAPI := &options.ServerAPIOptions{
		ServerAPIVersion:  options.ServerAPIVersion1,
		Strict:            new(bool),
		DeprecationErrors: new(bool),
	}
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	mongoClient = client
	return err
}
