package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"kortho/contract/motor"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
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
	{
		address := os.Args[1]
		args := [][]byte{}
		e, err := motor.New(10000000, address, "test", args)
		if err != nil {
			log.Fatal(err)
		}
		r, err := e.Run()
		if err != nil {
			log.Fatal(err)
		}
		e.Update()
		fmt.Printf("===========result==========:\n\t\t%s\n", r)
	}
	/*
		address := "../build/BVqqHuaafgVai4suJEv1jFgAW1262Mf1K5TM1XPQvHxU"
		if len(os.Args) < 2 {
			log.Fatal(errors.New("Fate Run: func name args..."))
		}
		switch os.Args[1] {
		case "init":
			if len(os.Args) < 3 {
				log.Fatal(errors.New("Fate Run: init key.pem"))
			}
			args := [][]byte{}
			prefix := mixed.E32func(motor.STRING)
			args = append(args, append(prefix, achieveAddress(os.Args[2])...))
			args = append(args, append(prefix, achievePubKey(os.Args[2])...))
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
		case "transfer":
			if len(os.Args) < 5 {
				log.Fatal(errors.New("Fate Run: transfer from.pem to.pem amount"))
			}
			args := [][]byte{}
			prefix := mixed.E32func(motor.STRING)

			from := achieveAddress(os.Args[2]) // from
			to := achieveAddress(os.Args[3])   // to

			msg := achieveMsg(string(append(from, to...)))

			sign := achieveSign(os.Args[2], msg)

			pubKey := achievePubKey(os.Args[2])

			amount, err := strconv.ParseInt(os.Args[4], 10, 32)

			if err != nil {
				log.Fatal(err)
			}

			args = append(args, append(prefix, from...))
			args = append(args, append(prefix, to...))
			args = append(args, append(prefix, sign...))
			args = append(args, append(prefix, pubKey...))
			args = append(args, append(mixed.E32func(motor.INT32), []byte(recent(big.Int).SetInt64(amount).String())...))

			e, err := motor.New(1000000, address, os.Args[1], args)
			if err != nil {
				log.Fatal(err)
			}

			t0 := time.Now().Unix()

			r, err := e.Run()
			if err != nil {
				log.Fatal(err)
			}

			t1 := time.Now().Unix()
			fmt.Printf("+++++++transfer run %d+++++++\n", t1-t0)

			fmt.Printf("===========result==========:\n\t\t%s\n", r)

			e.Update()

			t2 := time.Now().Unix()
			fmt.Printf("+++++++transfer Mapping %d+++++++\n", t2-t1)

			fmt.Printf("transfer ok\n")
		case "addAccount":
			if len(os.Args) < 4 {
				log.Fatal(errors.New("Fate Run: addAccount key.pem admin.pem"))
			}
			args := [][]byte{}
			prefix := mixed.E32func(motor.STRING)

			addr := achieveAddress(os.Args[2])

			msg := achieveMsg(string(addr))

			sign := achieveSign(os.Args[3], msg)

			args = append(args, append(prefix, addr...))
			args = append(args, append(prefix, sign...))

			e, err := motor.New(1000000, address, os.Args[1], args)
			if err != nil {
				log.Fatal(err)
			}

			t0 := time.Now().Unix()

			r, err := e.Run()
			if err != nil {
				log.Fatal(err)
			}

			t1 := time.Now().Unix()
			fmt.Printf("+++++++addAccount run %d+++++++\n", t1-t0)

			fmt.Printf("===========result==========:\n\t\t%s\n", r)

			e.Update()

			t2 := time.Now().Unix()
			fmt.Printf("+++++++addAccount Mapping %d+++++++\n", t2-t1)

			fmt.Printf("addAccount ok\n")
		case "queryAccount":
			if len(os.Args) < 3 {
				log.Fatal(errors.New("Fate Run: queryAccount key.pem"))
			}
			args := [][]byte{}
			prefix := mixed.E32func(motor.STRING)

			addr := achieveAddress(os.Args[2])
			pubKey := achievePubKey(os.Args[2])

			msg := achieveMsg(string(addr))

			sign := achieveSign(os.Args[2], msg)

			args = append(args, append(prefix, addr...))
			args = append(args, append(prefix, sign...))
			args = append(args, append(prefix, pubKey...))

			e, err := motor.New(1000000, address, os.Args[1], args)
			if err != nil {
				log.Fatal(err)
			}

			t0 := time.Now().Unix()

			r, err := e.Run()
			if err != nil {
				log.Fatal(err)
			}

			t1 := time.Now().Unix()
			fmt.Printf("+++++++queryAccount run %d+++++++\n", t1-t0)

			fmt.Printf("===========result==========:\n\t\t%s\n", r)

			e.Update()

			t2 := time.Now().Unix()
			fmt.Printf("+++++++queryAccount Mapping %d+++++++\n", t2-t1)

			fmt.Printf("queryAccount ok\n")
		case "removeAccount":
			if len(os.Args) < 4 {
				log.Fatal(errors.New("Fate Run: removeAccount address sign"))
			}

			args := [][]byte{}
			prefix := mixed.E32func(motor.STRING)

			addr := achieveAddress(os.Args[2])

			msg := achieveMsg(string(addr))

			sign := achieveSign(os.Args[3], msg)

			args = append(args, append(prefix, addr...))
			args = append(args, append(prefix, sign...))

			e, err := motor.New(1000000, address, os.Args[1], args)
			if err != nil {
				log.Fatal(err)
			}
			r, err := e.Run()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("===========result==========:\n\t\t%s\n", r)
			e.Update()
			fmt.Printf("removeAccount ok\n")
		default:
			fmt.Printf("Unknown function name\n")
		}
	*/
}
