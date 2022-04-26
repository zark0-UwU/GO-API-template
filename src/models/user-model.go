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
	Name     string             `bson:"name,omitempty" json:"name"`
	Role     string             `bson:"role,omitempty" json:"role"`
	RoleID   primitive.ObjectID `bson:"roleID,omitempty" json:"roleID"`
	// Tokens list is only used to be able to block the token later by placing said token onto the BlockedTokens list
	Tokens []string `bson:"tokens,omitempty" json:"tokens"`
	// Any attempt to use the tokens stored here, will be blocked
	BlockedTokens []string `bson:"blockedTokens,omitempty" json:"blockedTokens"`
}

//? The plurificated interfaces of the models are probably useless AAAND anoying
type Users interface {
	Create() *mongo.InsertOneResult
	ReadAll() []User
	CreateSingletonDBAndCollection()
}

// userdata viewable by anyone
type userPublic struct {
	Username string `bson:"username,omitempty" json:"username"`
	Role     string `bson:"role,omitempty" json:"role"`
}

// User data only shown to the user
type userPrivate struct {
	Username string `bson:"username,omitempty" json:"username"`
	Email    string `bson:"email,omitempty" json:"email"`
	Name     string `bson:"name,omitempty" json:"name"`
	Role     string `bson:"role,omitempty" json:"role"`
	//RoleID   primitive.ObjectID `bson:"roleID,omitempty" json:"roleID"`
	// Tokens list is only used to be able to block the token later by placing said token onto the BlockedTokens list
	Tokens []string `bson:"tokens,omitempty" json:"tokens"`
	// Any attempt to use the tokens stored here, will be blocked
	BlockedTokens []string `bson:"blockedTokens,omitempty" json:"blockedTokens"`
}

var UsersCollection *mongo.Collection

// this is the database where the collection is expected, could have multiple if necessary
var userModelDB *mongo.Database

// Returns the public/viewable user info
func (u User) Public() userPublic {
	return userPublic{
		Username: u.Username,
		Role:     u.Role,
	}
}

// Returns the user viewable info
func (u User) Private() userPrivate {
	return userPrivate{
		Username:      u.Username,
		Email:         u.Email,
		Name:          u.Name,
		Role:          u.Role,
		Tokens:        u.Tokens,
		BlockedTokens: u.BlockedTokens,
	}
}

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

// Returns the specific data about the role for the user from the DB
func (u User) RoleData() (Role, error) {
	filter := bson.D{
		{"_id", u.RoleID.Hex()},
	}
	res := RolesCollection.FindOne(services.Mongo.Context, filter)

	var role Role
	err := res.Decode(role)

	if err != nil {
		return role, err
	}

	return role, nil
}

// Sets the role id by searching it via the name stored on user.Role
func (u *User) SetRole() error {
	filter := bson.D{
		{"role", u.Role},
	}
	res := RolesCollection.FindOne(services.Mongo.Context, filter)

	var role Role
	err := res.Decode(role)

	if err == nil {
		u.RoleID = role.ID
	}
	return err
}
