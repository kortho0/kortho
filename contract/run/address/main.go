package main

import (
	"fmt"
	"log"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
)

func main() {
	for i := 0; i < 200; i++ {
		priv, err := sm2.GenerateKey()
		if err != nil {
			log.Fatal(err)
		}
		pubData := sm2.Compress(&priv.PublicKey)
		h := sm3.New()
		h.Write(pubData)
		hData := h.Sum(nil)
		sm2.WritePrivateKeytoPem(fmt.Sprintf("Key%d.pem", i), priv, nil)
		fmt.Printf("pubkey = %s\naddress = %s\n", base58.Encode(pubData), base58.Encode(hData))
	}
}
