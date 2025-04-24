from node:22-alpine as client

workdir /app

copy ./package.json .
copy ./package-lock.json .

copy ./pkgs ./pkgs
copy ./ts ./ts

copy ./Taskfile.yml .
copy ./tailwind.config.cjs .
copy ./main.css .
copy ./tsconfig.json .

run npm install -g @go-task/cli
run npm ci

run task build/tailwind
run task build/ts

from golang:1.24-alpine as server_binary

workdir /app

copy ./go.mod .
copy ./go.sum .

copy ./pkgs ./pkgs

copy ./main.go .

run go install github.com/a-h/templ/cmd/templ@v0.3.857
run templ generate

run go build .


from alpine

workdir /app

copy --from=server_binary /app/twau.app ./server
copy --from=client /app/static ./static

cmd ["./server"]
