package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00) // 0 ın 16 lık gosterımıdır
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey //eliptik eğrisi ile private key
	PublicKey  []byte
}

// Serialize, cüzdanı serileştirmek için özel bir yöntem
func (w *Wallet) Serialize() []byte {
	var content bytes.Buffer

	// Private key'in D, X ve Y değerlerini kaydet
	content.Write(w.PrivateKey.D.Bytes())
	content.Write(w.PrivateKey.PublicKey.X.Bytes())
	content.Write(w.PrivateKey.PublicKey.Y.Bytes())
	content.Write(w.PublicKey)

	return content.Bytes()
}

// DeserializeWallet, serileştirilmiş veriyi Wallet yapısına dönüştürür
func DeserializeWallet(data []byte) *Wallet {
	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = elliptic.P256()

	privateKey.D = new(big.Int).SetBytes(data[:32])
	privateKey.PublicKey.X = new(big.Int).SetBytes(data[32:64])
	privateKey.PublicKey.Y = new(big.Int).SetBytes(data[64:96])
	publicKey := data[96:]

	return &Wallet{
		PrivateKey: *privateKey,
		PublicKey:  publicKey,
	}
}

// NewKeyPair fonksiyonu, bir private ve public key olusturur
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
	return secondHash[:checksumLength]        // checksum kodu olusturulur
}

// Address fonksiyonu, bir adres olusturur
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)                // public key hash kodu olusturulur
	versionedHash := append([]byte{version}, pubHash...) // version ve public key hash kodu birleştirilir
	checksum := Checksum(versionedHash)                  // checksum kodu olusturulur
	fullHash := append(versionedHash, checksum...)       // versionedHash ve checksum kodu birleştirilir
	address := Base58Encode(fullHash)                    // adres olusturulur

	return address
}

/*

    Address		 	: 14LErwM2aHhdsDym6PKyutyG9ZSm51UHXc
	FullHash	 	: 002348bd9e7a51b7aba9766a7c62d502079020802bc6c767
	[version]   	: 00
	[pubKeyHash] 	: 2348bd9e7a51b7aba9766a7c62d50207902080
	[checksum]		: 2bc6c767

	address alınır ve addresi base58 ıle decode edılır ve pubKey elde edilir

	fullhasın ilk karakterını sokup alın bu surum karakterıdır burada versıon 00 oldu

	ardından pubkeyhash kısmınıda cıkarın ve elımızde checkSum kalır 1bc6c767

*/

// ValidateAddress fonksiyonu, bir adresin gecerli olup olmadıgını kontrol eder
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))                        // adresi byte dizisine dönüştürülür
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]      // checksum kodu alınır pubKeyHash[5:] 5. indeks den sona kadar oalnı alır
	version := pubKeyHash[0]                                           // version kodu alınır
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]        // version ve checksum kodu silinir
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...)) // checksum kodu olusturulur

	return bytes.Compare(actualChecksum, targetChecksum) == 0 // checksum kodu karsılastırılır
}
