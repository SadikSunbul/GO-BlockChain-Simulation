package blockchain

import (
	"bytes"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
)

type TxOutput struct { //transectıon cıktıları
	Value     int    //token degeri
	PublicKey []byte // public key hash
}

type TxInput struct { //transectıon girdileri
	ID        []byte //cıkısı referans eder
	Out       int    //cıkıs endexı  referans eder
	Signature []byte // imza
	PubKey    []byte // public key
}

// UsesKey fonksiyonu, TxInput yapısının bir public key hash kodunu kullanıp kullanmadığını kontrol eder.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	// TxInput içindeki PubKey bilgisinden elde edilen public key hash'i alır.
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	// Elde edilen public key hash'i, verilen pubKeyHash ile karşılaştırır ve eşit olup olmadığını kontrol eder.
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// Lock fonksiyonu, TxOutput yapısını belirtilen bir adresle kilitleyerek (lock) TxOutput'un PublicKey alanını ayarlar.
func (out *TxOutput) Lock(address []byte) {
	// Verilen adresi Base58Decode fonksiyonu ile byte dizisine dönüştürür.
	pubKeyHash := wallet.Base58Decode(address)

	// Dönüştürülen byte dizisinden, versiyon (version) ve checksum kodunu çıkarır.
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	// Elde edilen byte dizisini TxOutput'un PublicKey alanına atar, bu alanı public key hash kodu olarak ayarlar.
	out.PublicKey = pubKeyHash
}

// IsLockedWithKey fonksiyonu, TxOutput yapısının belirtilen bir public key hash koduyla kilidinin kontrol edilmesini sağlar.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	// TxOutput'un PublicKey alanını, verilen pubKeyHash ile karşılaştırır ve eşit olup olmadığını kontrol eder.
	return bytes.Compare(out.PublicKey, pubKeyHash) == 0
}

// NewTXOutput fonksiyonu, bir token değeri (value) ve bir adresi (address) alarak yeni bir TxOutput yapısı oluşturur.
func NewTXOutput(value int, address string) *TxOutput {
	// Yeni bir TxOutput yapısı oluşturulur.
	txo := &TxOutput{value, nil}

	// Verilen adresi byte dizisine dönüştürerek ve bu adresle TxOutput'u kilitleyerek (lock) PublicKey alanını ayarlar.
	txo.Lock([]byte(address))

	// Hazırlanan TxOutput yapısını döndürür.
	return txo
}
