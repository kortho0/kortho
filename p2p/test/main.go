// package main

// import (
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"kortho/p2p/node"
// 	"log"

// 	"github.com/valyala/fasthttp"
// )

// func main() {
// 	port := flag.Int("p", 7777, "port")
// 	name := flag.String("n", "a", "node's name")
// 	address := flag.String("a", "127.0.0.1", "node's address")
// 	member := flag.String("m", "127.0.0.1:7777", "joined node's address")
// 	httpPort := flag.Int("hp", 8888, "http's port")
// 	flag.Parse()
// 	n, err := node.New(*port, *name, *address, nil, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	go n.Run()
// 	if *name != "a" {
// 		fmt.Printf("join: %v\n", n.Join([]string{*member}))
// 	}
// 	h := func(ctx *fasthttp.RequestCtx) {
// 		switch string(ctx.Path()) {
// 		case "/get":
// 			get(n, ctx)
// 		case "/set":
// 			set(n, ctx)
// 		default:
// 			ctx.Error("unsupport path", fasthttp.StatusNotFound)
// 		}
// 	}
// 	srv := &fasthttp.Server{Handler: h}
// 	srv.ListenAndServe(fmt.Sprintf(":%v", *httpPort))
// }

// func get(n node.Node, ctx *fasthttp.RequestCtx) {
// 	var mp map[string]interface{}

// 	ctx.Response.SetStatusCode(200)
// 	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
// 	ctx.Response.Header.Set("Content-Type", "application/json")
// 	if err := json.Unmarshal(ctx.PostBody(), &mp); err != nil {
// 		ctx.Response.SetStatusCode(400)
// 		ctx.Write([]byte(err.Error()))
// 		return
// 	}
// 	k, err := getString(mp, "key")
// 	if err != nil {
// 		ctx.Response.SetStatusCode(400)
// 		ctx.Write([]byte(err.Error()))
// 		return
// 	}
// 	ctx.Write([]byte(n.Get(k)))
// }

// func set(n node.Node, ctx *fasthttp.RequestCtx) {
// 	var mp map[string]interface{}

// 	ctx.Response.SetStatusCode(200)
// 	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
// 	ctx.Response.Header.Set("Content-Type", "application/json")
// 	if err := json.Unmarshal(ctx.PostBody(), &mp); err != nil {
// 		ctx.Response.SetStatusCode(400)
// 		ctx.Write([]byte(err.Error()))
// 		return
// 	}
// 	k, err := getString(mp, "key")
// 	if err != nil {
// 		ctx.Response.SetStatusCode(400)
// 		ctx.Write([]byte(err.Error()))
// 		return
// 	}
// 	v, err := getString(mp, "value")
// 	if err != nil {
// 		ctx.Response.SetStatusCode(400)
// 		ctx.Write([]byte(err.Error()))
// 		return
// 	}
// 	n.Set(k, v)
// 	ctx.Write([]byte("ok"))
// }

// func getString(mp map[string]interface{}, k string) (string, error) {
// 	v, ok := mp[k]
// 	if !ok {
// 		return "", fmt.Errorf("'%s' not exist", k)
// 	}
// 	if _, ok := v.(string); !ok {
// 		return "", fmt.Errorf("'%s' not string", k)
// 	}
// 	return v.(string), nil
// }
