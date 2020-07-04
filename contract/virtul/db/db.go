package db

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func isExit(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func New(name string) (*db, error) {
	if !isExit(name) {
		if err := os.Mkdir(name, os.FileMode(0777)); err != nil {
			return nil, err
		}
	}
	return &db{name}, nil
}

func (a *db) Close() error {
	return nil
}

func (a *db) Del(k []byte) error {
	if a == nil || len(k) == 0 {
		return errors.New("Flash Db Del: Illegal Arguments")
	}
	return os.Remove(fmt.Sprintf("%s/%s.pg", a.name, string(k)))
}

func (a *db) Set(k, v []byte) error {
	var err error
	var fp *os.File

	if a == nil || len(v) != PAGE_SIZE {
		return errors.New("Flash Db Set: Illegal Arguments")
	}
	file := fmt.Sprintf("%s/%s.pg", a.name, string(k))
	if !isExit(file) {
		if fp, err = os.Create(file); err != nil {
			return fmt.Errorf("Flash Db Set: %v", err)
		}
	} else {
		if fp, err = os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, os.FileMode(0777)); err != nil {
			return fmt.Errorf("Flash Db Set: %v", err)
		}
	}
	if n, err := fp.Write(v); n != PAGE_SIZE {
		return fmt.Errorf("Flash Db Set: %v", err)
	}
	return fp.Close()
}

func (a *db) Get(k []byte) ([]byte, error) {
	var err error
	var fp *os.File

	if a == nil {
		return []byte{}, errors.New("Flash Db Get: Illegal Arguments")
	}
	file := fmt.Sprintf("%s/%s.pg", a.name, string(k))
	if !isExit(file) {
		return []byte{}, errors.New("Flash Db Get: Key Is Not Exist")
	}
	if fp, err = os.Open(file); err != nil {
		return []byte{}, fmt.Errorf("Flash Db Get: %v", err)
	}
	buf := make([]byte, PAGE_SIZE)
	if n, err := fp.Read(buf); n != PAGE_SIZE {
		return []byte{}, fmt.Errorf("Flash Db Get: %v", err)
	}
	return buf, fp.Close()
}

func (a *db) GetExecute() ([]byte, error) {
	if a == nil {
		return []byte{}, errors.New("Flash Db Get Exectue File: Illegal Arguments")
	}
	return ioutil.ReadFile(fmt.Sprintf("%s/ft", a.name))
}

func (a *db) SetExecute(ft []byte) error {
	if a == nil {
		return errors.New("Flash Db Set Exectue File: Illegal Arguments")
	}
	return ioutil.WriteFile(fmt.Sprintf("%s/ft", a.name), ft, os.FileMode(0777))
}
