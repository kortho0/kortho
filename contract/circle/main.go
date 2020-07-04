package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"

	"kortho/contract/mixed"
	"kortho/contract/motor"
)

func achieveAddress(key string) []byte {
	priv, err := sm2.ReadPrivateKeyFromPem(key, nil)
	if err != nil {
		log.Fatal(err)
	}
	pubData := sm2.Compress(&priv.PublicKey)
	h := sm3.New()
	h.Write(pubData)
	hData := h.Sum(nil)
	return []byte(base58.Encode(hData))
}

func achievePubKey(key string) []byte {
	priv, err := sm2.ReadPrivateKeyFromPem(key, nil)
	if err != nil {
		log.Fatal(err)
	}
	pubData := sm2.Compress(&priv.PublicKey)
	return []byte(base58.Encode(pubData))
}

func achieveMsg(address string) []byte {
	msg := []byte{}
	msg = append([]byte(address), []byte("1")...)
	return msg
}

func achieveSign(key string, msg []byte) []byte {
	priv, err := sm2.ReadPrivateKeyFromPem(key, nil)
	if err != nil {
		log.Fatal(err)
	}
	sign, err := priv.Sign(rand.Reader, msg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return []byte(base58.Encode(sign))
}

func main() {
	var e motor.Fate

	address := "../build/BVqqHuaafgVai4suJEv1jFgAW1262Mf1K5TM1XPQvHxU"
	if len(os.Args) < 2 {
		log.Fatal(errors.New("Fate Run: func name args..."))
	}
	dir, err := ioutil.ReadDir("address")
	if err != nil {
		log.Fatal(err)
	}
	switch os.Args[1] {
	case "init":
		if len(os.Args) < 2 {
			log.Fatal(errors.New("Fate Run: init key.pem"))
		}
		args := [][]byte{}
		prefix := mixed.E32func(motor.STRING)
		args = append(args, append(prefix, achieveAddress("admin.pem")...))
		args = append(args, append(prefix, achievePubKey("admin.pem")...))
		e, err := motor.New(10000, address, os.Args[1], args)
		if err != nil {
			log.Fatal(err)
		}

		t0 := time.Now().Unix()

		r, err := e.Run()
		if err != nil {
			log.Fatal(err)
		}

		t1 := time.Now().Unix()
		fmt.Printf("+++++++init run %d+++++++\n", t1-t0)

		fmt.Printf("===========result==========:\n\t\t%s\n", r)

		e.Update()
		t2 := time.Now().Unix()
		fmt.Printf("+++++++init Mapping %d+++++++\n", t2-t1)

		fmt.Printf("init ok\n")

	case "add":
		for _, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".pem")
			if ok {
				file := append([]byte("./address/"), []byte(fi.Name())...)

				args := [][]byte{}
				prefix := mixed.E32func(motor.STRING)
				addr := achieveAddress(string(file))
				msg := achieveMsg(string(addr))
				sign := achieveSign("admin.pem", msg)
				args = append(args, append(prefix, addr...))
				args = append(args, append(prefix, sign...))

				if e == nil {
					e, _ = motor.New(1000000, address, "addAccount", args)
				} else {
					e.Dup(1000000, "addAccount", args)
				}
				r, err := e.Run()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("===========result==========:\n\t\t%s\n", r)
				if e.Memory() > uint64(1024)*1024*1000 {
					fmt.Printf("=========================update======================\n")
					e.Update()
				}
			}
		}
		e.Update()

	case "trans":
		for _, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".pem")
			if ok {
				file := append([]byte("./address/"), []byte(fi.Name())...)
				args := [][]byte{}
				prefix := mixed.E32func(motor.STRING)

				fromKey := "admin.pem"

				from := achieveAddress(fromKey)    // from
				to := achieveAddress(string(file)) // to

				msg := achieveMsg(string(append(from, to...)))
				sign := achieveSign(fromKey, msg)

				pubKey := achievePubKey(fromKey)

				amount := recent(big.Int).SetInt64(1).String()

				args = append(args, append(prefix, from...))
				args = append(args, append(prefix, to...))
				args = append(args, append(prefix, sign...))
				args = append(args, append(prefix, pubKey...))
				args = append(args, append(mixed.E32func(motor.INT32), []byte(amount)...))

				if e == nil {
					e, _ = motor.New(1000000, address, "transfer", args)
				} else {
					e.Dup(1000000, "transfer", args)
				}
				r, err := e.Run()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("===========result %s %s==========:\n\t\t%s\n\n", file, amount, r)
				if e.Memory() > uint64(1024)*1024*1000 {
					fmt.Printf("=========================update======================\n")
					e.Update()
				}
			}
		}
		e.Update()

	case "query":
		for _, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".pem")
			if ok {
				file := append([]byte("./address/"), []byte(fi.Name())...)
				args := [][]byte{}
				prefix := mixed.E32func(motor.STRING)

				addr := achieveAddress(string(file))
				pubKey := achievePubKey(string(file))

				msg := achieveMsg(string(addr))

				sign := achieveSign(string(file), msg)

				args = append(args, append(prefix, addr...))
				args = append(args, append(prefix, sign...))
				args = append(args, append(prefix, pubKey...))

				if e == nil {
					e, _ = motor.New(1000000, address, "queryAccount", args)
				} else {
					e.Dup(1000000, "queryAccount", args)
				}

				r, err := e.Run()
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("===========result %s==========:\n\t\t%s\n\n", file, r)

				if e.Memory() > uint64(1024)*1024*500 {
					fmt.Printf("=========================update======================\n")
					e.Update()
				}
			}
		}
		e.Update()
	case "remove":
		for _, fi := range dir {
			ok := strings.HasSuffix(fi.Name(), ".pem")
			if ok {
				file := append([]byte("./address/"), []byte(fi.Name())...)
				args := [][]byte{}
				prefix := mixed.E32func(motor.STRING)

				addr := achieveAddress(string(file))

				msg := achieveMsg(string(addr))

				sign := achieveSign("admin.pem", msg)

				args = append(args, append(prefix, addr...))
				args = append(args, append(prefix, sign...))
				if e == nil {
					e, _ = motor.New(1000000, address, "removeAccount", args)
				} else {
					e.Dup(1000000, "removeAccount", args)
				}
				r, err := e.Run()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("run result: %s\n", r)

				if e.Memory() > uint64(1024)*1024*500 {
					fmt.Printf("=========================update======================\n")
					e.Update()
				}
			}
		}
		e.Update()
	default:
		fmt.Printf("Unknown function name\n")

	}
}
