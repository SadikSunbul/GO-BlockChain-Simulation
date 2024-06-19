package wallet

import (
	"github.com/mr-tron/base58"
	"log"
)

// Base58 fonksiyonu, verilen byte dizisini base58 koduna doğru dönüştürür
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input) // Byte dizisini base58 koduna dönüştürür
	return []byte(encode)          // Dönüştürülen kodu byte dizisine dönüştürür
}

// Base58Decode fonksiyonu, verilen base58 kodunu byte dizisine dönüştürür
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input)) // base58 kodunu byte dizisine dönüştürür
	if err != nil {
		log.Panic(err)
	}
	return decode
}

// 0 o l I + / bunlar sıfrelerde yoktur
