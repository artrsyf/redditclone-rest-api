package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	commentRepository "redditclone/pkg/comment/repository/mongo"
	postRepository "redditclone/pkg/post/repository/mongo"
	sessionRepository "redditclone/pkg/session/repository/redis"
	userRepository "redditclone/pkg/user/repository/mysql"

	commentDelivery "redditclone/pkg/comment/delivery"
	"redditclone/pkg/middleware"
	postDelivery "redditclone/pkg/post/delivery"
	userDelivery "redditclone/pkg/user/delivery"
	"redditclone/tools"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

const configPath = "config.yaml"

type Config struct {
	StaticRoot string `yaml:"STATIC_ROOT"`
	Port       int    `yaml:"PORT"`
}

var AppConfig *Config

func main() {
	tools.Init()

	err := godotenv.Load()
	if err != nil {
		tools.Logger.Fatal("error loading .env file:", err)
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		tools.Logger.Fatal("error opening config file:", err)
	}

	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		tools.Logger.Fatal("error reading config file:", err)
	}

	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	mysqlDSN += "&charset=utf8"
	mysqlDSN += "&interpolateParams=true"

	mysqlConnect, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		panic(err)
	}

	mysqlConnect.SetConnMaxLifetime(time.Minute * 3)
	mysqlConnect.SetMaxOpenConns(10)
	mysqlConnect.SetMaxIdleConns(10)

	ctx := context.Background()
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/?maxPoolSize=10",
		os.Getenv("MONGODB_USER"),
		os.Getenv("MONGODB_PASSWORD"),
		os.Getenv("MONGODB_HOST"),
		os.Getenv("MONGODB_PORT"),
	)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	mongoConnect, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}
	mongoDB := mongoConnect.Database(os.Getenv("MONGODB_DATABASE"))
	postsCollection := mongoDB.Collection("posts")
	commentsCollection := mongoDB.Collection("comments")

	redisURL := fmt.Sprintf("redis://user:@%s:%s/%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_DATABASE"),
	)
	redisConn, err := redis.DialURL(redisURL)
	if err != nil {
		panic(err)
	}

	defer func() {
		configFile.Close()

		if err := mongoConnect.Disconnect(ctx); err != nil {
			panic(err)
		}

		if err = mysqlConnect.Close(); err != nil {
			panic(err)
		}

		if err = redisConn.Close(); err != nil {
			panic(err)
		}
	}()

	router := mux.NewRouter()

	userRepo := userRepository.NewUserMySqlRepo(mysqlConnect)
	sessionRepo := sessionRepository.NewSessionRedisManager(redisConn)
	postRepo := postRepository.NewPostMongoDBMemoryRepo(postsCollection)
	commentRepo := commentRepository.NewCommentMongoDBRepository(commentsCollection)

	postHandler := postDelivery.PostHandler{
		CommentRepo: commentRepo,
		PostRepo:    postRepo,
		UserRepo:    userRepo,
	}

	commentHandler := commentDelivery.CommentHandler{
		CommentRepo: commentRepo,
		PostRepo:    postRepo,
		UserRepo:    userRepo,
	}

	authHandler := userDelivery.UserHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}

	fileServer := http.FileServer(http.Dir(AppConfig.StaticRoot))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	router.Handle("/api/post/{postID}/upvote",
		middleware.ValidateJWTToken(
			sessionRepo,
			http.HandlerFunc(postHandler.Upvote))).Methods("GET")

	router.Handle("/api/post/{postID}/downvote",
		middleware.ValidateJWTToken(
			sessionRepo,
			http.HandlerFunc(postHandler.Downvote))).Methods("GET")

	router.Handle("/api/post/{postID}/unvote",
		middleware.ValidateJWTToken(
			sessionRepo,
			http.HandlerFunc(postHandler.Unvote))).Methods("GET")

	router.HandleFunc("/api/posts/", postHandler.Index).Methods("GET")

	router.HandleFunc("/api/posts/{category}", postHandler.IndexByCategory).Methods("GET")

	router.HandleFunc("/api/post/{id}", postHandler.GetPost).Methods("GET")

	router.Handle("/api/posts",
		middleware.ValidateContentType(
			middleware.ValidateJWTToken(
				sessionRepo,
				http.HandlerFunc(postHandler.Create)))).Methods("POST")

	router.Handle("/api/post/{postID}",
		middleware.ValidateContentType(
			middleware.ValidateJWTToken(
				sessionRepo,
				http.HandlerFunc(commentHandler.Create)))).Methods("POST")

	router.Handle("/api/post/{postID}",
		middleware.ValidateContentType(
			middleware.ValidateJWTToken(
				sessionRepo,
				http.HandlerFunc(postHandler.Delete)))).Methods("DELETE")

	router.Handle("/api/post/{postID}/{commentID}",
		middleware.ValidateContentType(
			middleware.ValidateJWTToken(
				sessionRepo,
				http.HandlerFunc(commentHandler.Delete)))).Methods("DELETE")

	router.HandleFunc("/api/user/{username}", postHandler.IndexByUser).Methods("GET")

	router.Handle("/api/login", middleware.ValidateContentType(
		http.HandlerFunc(authHandler.Login))).Methods("POST")

	router.Handle("/api/register", middleware.ValidateContentType(
		http.HandlerFunc(authHandler.Signup))).Methods("POST")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(AppConfig.StaticRoot + "/html/index.html")
		if err != nil {
			tools.Logger.Fatal("error due parsing index.html:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			tools.Logger.Fatal("error due executing index.html:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	tools.Logger.Printf("starting server at http://127.0.0.1:%d", AppConfig.Port)
	tools.Logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", AppConfig.Port), router))
}
