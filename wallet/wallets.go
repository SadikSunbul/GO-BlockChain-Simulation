package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "./tmp/wallets.data" //Badgerı kullanmıycaz buradakı cuzdsanı saklamak ıcın

type Wallets struct {
	Wallets map[string]*Wallet
}

// CreateWallets fonksiyonu, bir Wallets nesnesi olusturur
func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}                       // wallet nesnesi olusturulur
	wallets.Wallets = make(map[string]*Wallet) // wallet nesnesi olusturulur

	err := wallets.LoadFile() // wallets dosyası okunur

	return &wallets, err // wallets nesnesi döndürülür
}

// GetWallet fonksiyonu, bir Wallet nesnesini döndürür
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address] // wallet nesnesi döndürülür
}

// AddWallet fonksiyonu, bir Wallet nesnesi ekler
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()                         // wallet nesnesi olusturulur
	address := fmt.Sprintf("%s", wallet.Address()) // wallet adresi stringe dönüştürülür
	ws.Wallets[address] = wallet                   // wallet nesnesi eklenir
	return address                                 // wallet adresi döndürülür
}

// GetAllAddress fonksiyonu, tüm wallet adreslerini döndürür
func (ws *Wallets) GetAllAddress() []string {
	var addresses []string            // adreslerin listesi olusturulur
	for address := range ws.Wallets { // tüm wallet adreslerini döndürür
		addresses = append(addresses, address) // adreslerin listesi eklenir
	}
	return addresses // adreslerin listesi döndürülür
}

// LoadFile fonksiyonu, dosya okunur
func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) { // dosya yoksa
		return err // hata döndürür
	}

	var wallets Wallets // wallet nesnesi olusturulur

	fileContent, err := os.ReadFile(walletFile) // dosya okunur

	if err != nil {
		log.Panic(err)
	} // hata döndürür
	gob.Register(elliptic.P256())                           // elliptic nesnesi olusturulur
	decoder := gob.NewDecoder(bytes.NewReader(fileContent)) // decoder nesnesi olusturulur
	err = decoder.Decode(&wallets)                          // decoder ile dosya okunur

	if err != nil {
		log.Panic(err) // hata döndürür
	}
	ws.Wallets = wallets.Wallets // wallet nesnesi olusturulur

	return nil // hata yok
}

// SaveFile fonksiyonu, dosya kaydedilir
func (ws *Wallets) SaveFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())       // elliptic nesnesi olusturulur
	encoder := gob.NewEncoder(&content) // encoder nesnesi oluşturulur
	err := encoder.Encode(ws)           // encoder ile dosya kaydedilir
	if err != nil {
		log.Panic(err)
	}
	err = os.WriteFile(walletFile, content.Bytes(), 0644) // dosya kaydedilir
	if err != nil {
		log.Panic(err)
	}
}
