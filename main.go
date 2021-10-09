package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MONGODB_URI = "mongodb://localhost:27017"
const COLLECTION_NAME = "testusers"
const DB_NAME = "testdb"

var collection *mongo.Collection

func main() {
	InitMongoDB()

	router := gin.Default()
	api := router.Group("/api")
	users := api.Group(("/users"))

	users.GET("/:size", GetUsers)
	users.POST("/insertUser", InsertUser)
	users.DELETE("/:uid/delete", DeleteUser)
	users.PUT("/:uid/update", UpdateUser)
	router.Run()
}

type User struct {
	UID           string  `bson:"_uid" json:"_uid"`
	ID            int     `bson:"id" json:"id"`
	IsActive      bool    `bson:"isActive" json:"isActive"`
	Balance       string  `bson:"balance" json:"balance"`
	Picture       string  `bson:"picture" json:"picture"`
	Age           int     `bson:"age" json:"age"`
	Name          string  `bson:"name" json:"name"`
	Gender        string  `bson:"gender" json:"gender"`
	Company       string  `bson:"company" json:"company"`
	Email         string  `bson:"email" json:"email"`
	Phone         string  `bson:"phone" json:"phone"`
	Address       string  `bson:"address" json:"address"`
	About         string  `bson:"about" json:"about"`
	Registered    string  `bson:"registered" json:"registered"`
	Latitude      float64 `bson:"latitude" json:"latitude"`
	Longitude     float64 `bson:"longitude" json:"longitude"`
	FavoriteFruit string  `bson:"favoriteFruit" json:"favoriteFruit"`
}

func InitMongoDB() {
	newClient, err := mongo.NewClient(options.Client().ApplyURI(MONGODB_URI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = newClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = newClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = newClient.Database(DB_NAME).Collection(COLLECTION_NAME)
}

func InsertUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}

	ctx := context.Background()
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func GetUsers(c *gin.Context) {
	sizeParam := c.Param("size")
	size, err := strconv.ParseInt(sizeParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx := context.Background()
	var users []User

	findOptions := options.Find()
	findOptions.SetLimit(size)
	searchResult, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	defer searchResult.Close(ctx)

	if err = searchResult.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func DeleteUser(c *gin.Context) {
	userUid := c.Param("uid")

	ctx := context.Background()
	var deletedUser User
	err := collection.FindOneAndDelete(ctx, bson.M{"_uid": userUid}).Decode(&deletedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotModified, gin.H{"ERROR": err.Error()})
			return

		}
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deletedUser)
}

func UpdateUser(c *gin.Context) {
	var user User
	userUid := c.Param("uid")

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}

	byteUser, err := bson.Marshal(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	var bUser bson.M
	if err = bson.Unmarshal(byteUser, &bUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	ctx := context.Background()
	var updatedUser User

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.M{"_uid": userUid}
	update := bson.D{{Key: "$set", Value: user}}
	err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotModified, gin.H{"ERROR": err.Error()})
			return

		}
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}
