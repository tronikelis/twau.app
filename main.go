package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"twau.app/pkgs/game_state"
	"twau.app/pkgs/server/controllers"
	"twau.app/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

type Env struct {
	Port       int    `env:"PORT"`
	Secret     []byte `env:"SECRET_KEY"`
	Production bool   `env:"PRODUCTION"`
}

func NewEnv() (*Env, error) {
	env := &Env{}

	val := reflect.ValueOf(env).Elem()
	for i := range val.NumField() {
		tag := val.Type().Field(i).Tag.Get("env")
		field := val.Field(i)

		str := os.Getenv(tag)
		if str == "" {
			return nil, fmt.Errorf("Env %s is empty", tag)
		}

		switch field.Interface().(type) {
		case int:
			int, err := strconv.Atoi(str)
			if err != nil {
				return nil, err
			}
			field.Set(reflect.ValueOf(int))
		case []byte:
			field.Set(reflect.ValueOf([]byte(str)))
		case string:
			field.Set(reflect.ValueOf(str))
		case bool:
			field.Set(reflect.ValueOf(str == "true"))
		}
	}

	return env, nil
}

func main() {
	env, err := NewEnv()
	if err != nil {
		log.Fatal(err)
	}

	server := maruchi.NewServer()

	server.Group("").
		Middleware(func(ctx maruchi.ReqContext, next maruchi.Handler) {
			ctx.Writer().Header().Set("cache-control", "public, max-age=31536000")
			next(ctx)
		}).
		Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	server.Middleware(func(ctx maruchi.ReqContext, next maruchi.Handler) {
		ctx.Writer().Header().Set("content-type", "text/html; charset=utf-8")
		next(ctx)
	})

	server.Middleware(req.MiddlewareReqContext(game_state.NewRooms(), env.Secret))

	controllers.Register(server)

	errChan := make(chan error)
	go func() {
		host := "localhost"
		if env.Production {
			host = "0.0.0.0"
		}
		errChan <- http.ListenAndServe(fmt.Sprintf("%s:%d", host, env.Port), server.ServeMux())
	}()

	log.Println("listening on", env.Port)
	log.Fatal(<-errChan)
}
