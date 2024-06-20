package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const (
	chechksumLength = 4
	version         = byte(0x00) // 0 ın 16 lık gosterımıdır
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey //eliptik eğrisi ile private key
	PublickKey []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256() //kullanıcagımız eliptik tipi burdakı sıfreler 256 byte olur

	private, err := ecdsa.GenerateKey(curve, rand.Reader) //rastgele sayı uretıcısı ıle bırlıkte uratebılrrıız artık
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) //public key olusturulur
	return *private, pubKey
}

// MakeWallet fonksiyonu, bir Wallet nesnesi olusturur
func MakeWallet() *Wallet {
	private, public := NewKeyPair()   // private ve public key olusturulur
	wallet := Wallet{private, public} // wallet nesnesi olusturulur
	return &wallet
}

// PublicKeyHash fonksiyonu, bir public key hash kodu olusturur
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)   // public key hash kodu olusturulur
	hasher := ripemd160.New()          // hasher nesnesi olusturulur
	_, err := hasher.Write(pubHash[:]) // public key hash kodu yazılır
	if err != nil {
		log.Panic(err)
	}
	pubicRipMD := hasher.Sum(nil) // public key hash kodu olusturulur
	return pubicRipMD
}

// Checksum fonksiyonu, bir checksum kodu olusturur
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)       // payload hash kodu olusturulur
	secondHash := sha256.Sum256(firstHash[:]) // firstHash hash kodu olusturulur
	return secondHash[:chechksumLength]       // checksum kodu olusturulur
}

// Address fonksiyonu, bir adres olusturur
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublickKey)               // public key hash kodu olusturulur
	versionedHash := append([]byte{version}, pubHash...) // version ve public key hash kodu birleştirilir
	checksum := Checksum(versionedHash)                  // checksum kodu olusturulur
	fullHash := append(versionedHash, checksum...)       // versionedHash ve checksum kodu birleştirilir
	address := Base58Encode(fullHash)                    // adres olusturulur

	return address
}
