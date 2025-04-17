package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"word-amongus-game/pkgs/game_state"
	"word-amongus-game/pkgs/server/controllers"
	"word-amongus-game/pkgs/server/req"

	"github.com/tronikelis/maruchi"
)

type Env struct {
	Port   int    `env:"PORT"`
	Secret []byte `env:"SECRET_KEY"`
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
			// if production {
			// 	ctx.Header().Set("cache-control", "public, max-age=31536000")
			// } else {
			ctx.Writer().Header().Set("cache-control", "no-cache, no-store, must-revalidate")
			// }
			next(ctx)
		}).
		Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	server.Middleware(req.MiddlewareReqContext(game_state.NewRooms(), env.Secret))

	controllers.Register(server)

	errChan := make(chan error)
	go func() {
		errChan <- http.ListenAndServe(fmt.Sprintf("localhost:%d", env.Port), server.ServeMux())
	}()

	log.Println("listening on", env.Port)

	panic(<-errChan)
}
