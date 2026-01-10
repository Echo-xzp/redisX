package command

import (
	"bytes"
	"fmt"
	"redisx/internal/storage"
)

// INCR key
func Incr(store *storage.Storage, args []string) ([]byte, error) {
	if len(args) < 1 {
		return []byte("-ERR wrong number of arguments for 'INCR' command\r\n"), nil
	}
	n, err := store.IncrBy(args[0], 1)
	if err != nil {
		return []byte("-ERR value is not an integer or out of range\r\n"), nil
	}
	return []byte(fmt.Sprintf(":%d\r\n", n)), nil
}

// MGET key [key ...]
func MGet(store *storage.Storage, args []string) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, k := range args {
		if v, ok := store.Get(k); ok {
			b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		} else {
			b.WriteString("$-1\r\n")
		}
	}
	return b.Bytes(), nil
}

// PERSIST key
func Persist(store *storage.Storage, args []string) ([]byte, error) {
	if len(args) < 1 {
		return []byte("-ERR wrong number of arguments for 'PERSIST' command\r\n"), nil
	}
	if store.Persist(args[0]) {
		return []byte(":1\r\n"), nil
	}
	return []byte(":0\r\n"), nil
}
