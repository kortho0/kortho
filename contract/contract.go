package contract

import (
	"bytes"
	"io/ioutil"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm3"
)

func codeAddress(code []byte) string {
	h := sm3.New()
	h.Write(code)
	name := h.Sum(nil)
	return base58.Encode(name)
}

func Motor(name string) (string, error) {
	if code, err := ioutil.ReadFile(name); err == nil {
		fp := NewParser(bytes.NewReader(code))
		if err = fp.Parser(); err != nil {
			return "", err
		}
		fg := NewGenerator(codeAddress(code))
		if err = fg.Generate(fp.ns); err != nil {
			return "", err
		}
		return codeAddress(code), nil
	} else {
		return "", err
	}
}
