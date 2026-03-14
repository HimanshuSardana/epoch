package main

import (
	"crypto/sha1"
	"fmt"
)

func hashObject(fileName string) {
	header := fmt.Sprintf("blob %d\x00", len(fileName))
	content := header + fileName

	hash := sha1.New()
	hash.Write([]byte(content))
	result := hash.Sum(nil)

	fmt.Printf("%x\n", result)
}
