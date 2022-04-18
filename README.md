# Go API template
## Using/implementing
 - **Fiber**    
    > For performant routing and many other things
 - **JWT**      
    > For user authentication
 - **MongoDB**  
    > As the main database
 - **godotenv** 
    > For local development
## Introduction/vision:

Fiber is a quickly growing super-fast backend framework, mainly written in golang with performance in mind. 

The intent of this repo is to create a **performant, comprenhensible, mantainable and production ready template** with nothing in it but the basics, so you do not have to spend time removing code or strugeling to understand any complicated code structure

The only thing it handles by default is a sesion management system.

## Future features:

- Documenting using [fiber-swagger](https://github.com/arsmn/fiber-swagger)
- Performance measurement using [otelfiber](https://github.com/gofiber/contrib/tree/main/otelfiber)

## Project Structure:

```
.
├── src/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── users/
│   │       ├── auth.go
│   │       └── user.go
│   ├── helpers/
│   │   └── .
│   ├── loaders/
│   │   ├── fiber.go
│   │   └── mongo.go
│   ├── middlewares/
│   │   └── auth.go
│   ├── models/
│   │   └── user-model.go
│   ├── routes/
│   │   ├── appRouter.go
│   │   ├── auth.go
│   │   └── users.go
│   ├── services/
│   │   └── mongo.go
│   ├── utils/
│   │   └── .
│   └── main.go
├── .env
├── app.go
├── go.mod
├── go.sum
├── README.md
└── sample.env
```

I wanted to contain all the source code under a sub folder, so I only use app.go to execute a Start() function in src/main.go

## Setup:

### If building from source:

1. Once this repository is cloned and golang is installed in the system, navigate to this directory and run

```
go mod download
```

2. Once the dependencies are downloaded, using sample.env as reference either create a file called creds.env with the same keys or directly configure same keys as environment variables.
3. After the configuration and ensuring that the db is operational, run either

```
go build # to get the executable to run

or

go run main.go
```

## Announcement:

If you find this repository useful to you in any way, do help me by starring this repository and let your friends know about this repo. I'll try my best to keep it updated.
