package models

import (
	"fmt"

	"GO-API-template/src/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username,omitempty" json:"username"`
	Email    string             `bson:"email,omitempty" json:"email"`
	Password string             `bson:"password,omitempty" json:"password"`
	FullName string             `bson:"fullName,omitempty" json:"fullName"`
	Role     string             `bson:"role,omitempty" json:"role"`
}

//? The plurificated interfaces of the models are probably useless AAAND anoying
type Users interface {
	Create() *mongo.InsertOneResult
	ReadAll() []User
	CreateSingletonDBAndCollection()
}

var UsersCollection *mongo.Collection

// this is the database where the collection is expected, could have multiple if necessary
var userModelDB *mongo.Database

func (u User) CreateSingletonDBAndCollection() {
	if userModelDB == nil {
		userModelDB = services.Mongo.DBs["mainDB"]
	}
	if UsersCollection == nil {
		UsersCollection = userModelDB.Collection("users")
	}
}

func (u User) Create() (*mongo.InsertOneResult, error) {
	u.CreateSingletonDBAndCollection()

	insertedRes, err := UsersCollection.InsertOne(services.Mongo.Context, u)
	if err != nil {
		fmt.Println(err)
	}

	return insertedRes, err
}

func (u User) ReadAll() []User {
	u.CreateSingletonDBAndCollection()

	filter := bson.D{
		//{"name", "uwu"}, // this works, Try it!
	}

	currsor, err := UsersCollection.Find(services.Mongo.Context, filter)
	if err != nil {
		panic(err)
	}
	defer currsor.Close(services.Mongo.Context)

	var users []User
	currsor.All(services.Mongo.Context, &users)

	return users
}

//TODO: Get next document in collection every time the function is called
func (u User) ReadOne() []bson.M { //? maybe shouldn't make this function
	u.CreateSingletonDBAndCollection()

	filter := bson.D{
		//{"name", "uwu"}, // this works, Try it!
	}

	currsor, err := UsersCollection.Find(services.Mongo.Context, filter)
	if err != nil {
		panic(err)
	}

	var kao []bson.M

	// TODO: check wether cursor is exhausted or fromStart == true , if is create new cursor and iterate that
	currsor.All(services.Mongo.Context, &kao)

	return kao
}
