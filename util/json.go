package util

import (
	"encoding/json"
	"os"
)

func JSONMarshalFile(filename string, v interface{}, perm os.FileMode) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(v)
}

func JSONMarshalIndentFile(filename string, v interface{}, prefix, indent string, perm os.FileMode) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return err
	}
	file.Write(b)

	return nil
}

func JSONUnmarshalFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	return dec.Decode(v)
}
