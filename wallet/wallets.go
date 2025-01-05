package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "./tmp/wallets_%s.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func init() {
	gob.Register(elliptic.P256())
}

// CreateWallets fonksiyonu, bir Wallets nesnesi olusturur
func CreateWallets(nodeId string) (*Wallets, error) {
	wallets := Wallets{}                       // wallet nesnesi olusturulur
	wallets.Wallets = make(map[string]*Wallet) // wallet nesnesi olusturulur

	err := wallets.LoadFile(nodeId) // wallets dosyası okunur

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

// LoadFile fonksiyonu, cüzdanları dosyadan yükler
func (ws *Wallets) LoadFile(nodeId string) error {
	walletFile := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var walletsData map[string][]byte

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&walletsData)
	if err != nil {
		return err
	}

	wallets := make(map[string]*Wallet)
	for addr, data := range walletsData {
		wallets[addr] = DeserializeWallet(data)
	}

	ws.Wallets = wallets
	return nil
}

// SaveFile fonksiyonu, cüzdanları dosyaya kaydeder
func (ws *Wallets) SaveFile(nodeId string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeId)

	var walletsData = make(map[string][]byte)
	for addr, wallet := range ws.Wallets {
		walletsData[addr] = wallet.Serialize()
	}

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(walletsData)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
