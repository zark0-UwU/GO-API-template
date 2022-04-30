package models

import (
	"context"
	"fmt"

	"GO-API-template/src/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Role struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	// Role string
	Role string `bson:"role,omitempty" json:"role"`
	// Integer to indicate the permissons level, the lower, the higer perms.
	// 0 is full acess, this is only used to compare roles
	Level int `bson:"level,omitempty" json:"level"`
	// Permissons object, this is used when performing a specific action
	Permissons Permissons `bson:"Permissons,omitempty" json:"Permissons"`
}

// Theese permissons are checked every time a user wants to make an operation restricted to certain roles
type Permissons struct { //? shuld have a FullAcess flag?
	ReadUsers  bool `bson:"readUsers,omitempty" json:"readUsers"`
	UsersAdmin bool `bson:"usersAdmin,omitempty" json:"usersAdmin"`
	ReadRoles  bool `bson:"readRoles,omitempty" json:"readRoles"`
	RolesAdmin bool `bson:"rolesAdmin,omitempty" json:"rolesAdmin"`
}

//? The plurificated interfaces of the models are probably useless AAAND anoying
type Roles interface {
	Create() *mongo.InsertOneResult
	ReadAll() []User
	CreateSingletonDBAndCollection()
}

var RolesCollection *mongo.Collection

// this is the database where the collection is expected, could have multiple if necessary
var rolesModelDB *mongo.Database

func (r Role) CreateSingletonDBAndCollection() {
	if rolesModelDB == nil {
		rolesModelDB = services.Mongo.DBs["mainDB"]
	}
	if RolesCollection == nil {
		RolesCollection = rolesModelDB.Collection("users")
	}
}

func (r Role) Create() (*mongo.InsertOneResult, error) {
	r.CreateSingletonDBAndCollection()

	insertedRes, err := RolesCollection.InsertOne(services.Mongo.Context, r)
	if err != nil {
		fmt.Println(err)
	}

	return insertedRes, err
}

func (r Role) ReadAll() []User {
	r.CreateSingletonDBAndCollection()

	filter := bson.D{
		//{"name", "uwu"}, // this works, Try it!
	}

	currsor, err := RolesCollection.Find(services.Mongo.Context, filter)
	if err != nil {
		panic(err)
	}
	defer currsor.Close(services.Mongo.Context)

	var users []User
	currsor.All(services.Mongo.Context, &users)

	return users
}

func (r *Role) Fill(identity string, id, role bool) error {
	var fields []bson.D
	if id {
		id, err := primitive.ObjectIDFromHex(identity)
		if err == nil {
			fields = append(fields, bson.D{{"_id", id}})
		}
	}
	if role {
		fields = append(fields, bson.D{{"role", identity}})
	}

	filter := bson.D{{"$or", fields}}

	res := RolesCollection.FindOne(context.Background(), filter)
	if err := res.Err(); err != nil {
		return err
	}
	res.Decode(r)
	return nil
}
