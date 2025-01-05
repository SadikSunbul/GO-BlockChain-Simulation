package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
)

func init() {
	gob.Register(elliptic.P256())
}

type Transaction struct {
	ID      []byte     //transectıon hası
	Inputs  []TxInput  //bu transectıondakı ınputlar
	Outputs []TxOutput //bu transectıondakı outputlar
}

// CoinbaseTx fonksiyonu, bir coinbase transaction oluşturur.
func CoinbaseTx(to, data string) *Transaction {
	if data == "" { //data boş ise gir
		randData := make([]byte, 24)  //data 24 byte'lık bir diziye dönüştür
		_, err := rand.Read(randData) //rastgele sayı uretıcısı ile diziye dönüştür (diziyi doldur)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData) // diziyi stringe doğru dönüştür

	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)} //hıcbır cıktıya referabs vermez ,cıkıs endexi -1 aynı referans yok , sadce data mesajı vardır
	txout := NewTXOutput(20, to)                     //100 tokeni to ya gonderırı

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}} //transectıonı olustururuz
	tx.ID = tx.Hash()                                           //Transectıon hashini olustururuz                                           //Transectıon Id sını olustururuz
	return &tx
}

// NewTransaction, belirtilen bir adresten başka bir adrese belirtilen miktar token transferi yapacak yeni bir işlem oluşturur.
func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := fmt.Sprintf("%s", w.Address())

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)
	return transaction
}

/*
IsCoinbase fonksiyonu, bir işlemin coinbase işlemi olup olmadığını belirler. Coinbase işlemi, yeni bir bloğun ödülünü
alan ilk işlemdir ve genellikle tek bir girişe sahiptir. Bu fonksiyon, giriş sayısını, girişin kimlik uzunluğunu ve
çıkışını kontrol ederek bir işlemin coinbase olup olmadığını belirler.
*/
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
	// Bu fonksiyon, bir işlemin coinbase işlemi olup olmadığını kontrol eder.
	// Bir coinbase işlemi sadece bir girişe sahiptir.
	// Bu girişin işlem kimliği (ID) uzunluğu 0'a eşit olmalıdır ve çıkış (Out) -1 olmalıdır.
}

// Serialize fonksiyonu, bir Transaction yapısını serileştirir (encode eder) ve byte dizisi olarak döndürür.
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer        // Yeni bir bytes.Buffer oluşturulur
	enc := gob.NewEncoder(&encoded) // gob (Go's binary serialization format) ile encode edici oluşturulur

	err := enc.Encode(tx) // Transaction yapısını encode eder
	if err != nil {
		log.Panic(err) // Hata durumunda hata mesajı gösterir ve işlemi sonlandırır
	}
	return encoded.Bytes() // Encode edilmiş veriyi byte dizisi olarak döndürür
}

// Hash fonksiyonu, bir Transaction yapısının hash değerini hesaplar ve byte dizisi olarak döndürür.
func (tx *Transaction) Hash() []byte {
	var hash [32]byte    // 32 byte'lık bir hash dizisi oluşturulur
	txCopy := *tx        // Transaction yapısının bir kopyası oluşturulur
	txCopy.ID = []byte{} // ID alanı temizlenir (boş byte dizisi atanır)

	hash = sha256.Sum256(txCopy.Serialize()) // Transaction yapısının serialize edilmiş halinin SHA-256 hash'ini hesaplar

	return hash[:] // Hesaplanan hash değerini byte dizisi olarak döndürür
}

// Sign fonksiyonu, bir Transaction yapısını imzalar.
// İmzalamak için verilen private anahtar (privKey) kullanılır ve işlemi daha önce yapılmış olan işlemlerle (prevTXs) ilişkilendirir.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() { // Eğer işlem bir coinbase işlemi ise (ödül işlemi ise)
		return // İşlem yapma, çünkü coinbase işlemleri imzalanmaz
	}

	for _, in := range tx.Inputs {
		// İşlemdeki her girdi için önceki işlem kontrolü yapılır
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct") // Önceki işlem doğruluğu sağlanmazsa hata ver ve işlemi sonlandır
		}
	}

	// İşlem yapısının bir kopyası oluşturulur ve gerekli alanlar temizlenir
	txCopy := tx.TrimmedCopy()

	// İşlemdeki her girdi için imzalama işlemi yapılır
	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]                  // Girdinin önceki işlem verisini alır
		txCopy.Inputs[inId].Signature = nil                           // İmza alanı temizlenir
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PublicKey // Girdiye ait PublicKey ayarlanır
		txCopy.ID = txCopy.Hash()                                     // İşlemin hash değeri hesaplanır
		txCopy.Inputs[inId].PubKey = nil                              // PublicKey alanı temizlenir (güvenlik amacıyla)

		// İşlemi imzalar
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID) // ECDSA algoritması kullanarak işlemi imzalar
		Handle(err)                                               // Hata durumunda işlemi ele alır
		signature := append(r.Bytes(), s.Bytes()...)              // İmza değerleri birleştirilir

		tx.Inputs[inId].Signature = signature // İşlemdeki girdiye imzayı ekler
	}
}

// Verify fonksiyonu, bir Transaction yapısının geçerliliğini doğrular.
// Geçerlilik kontrolü için verilen önceki işlemler haritası (prevTXs) kullanılır.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() { // Eğer işlem bir coinbase işlemi ise
		return true // Coinbase işlemleri doğrudur (her zaman geçerli)
	}

	// İşlemdeki her girdi için önceki işlem doğruluğu kontrol edilir
	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction not correct") // Önceki işlem doğruluğu sağlanmazsa hata verir ve işlemi sonlandırır
		}
	}

	// İşlem yapısının bir kopyası oluşturulur ve gerekli alanlar temizlenir
	txCopy := tx.TrimmedCopy()

	// ECDSA P256 eğrisi kullanılır
	curve := elliptic.P256()

	// Her girdi için imza doğrulaması yapılır
	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]                  // Girdinin önceki işlem verisini alır
		txCopy.Inputs[inId].Signature = nil                           // İmza alanı temizlenir
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PublicKey // Girdiye ait PublicKey ayarlanır
		txCopy.ID = txCopy.Hash()                                     // İşlemin hash değeri hesaplanır
		txCopy.Inputs[inId].PubKey = nil                              // PublicKey alanı temizlenir (güvenlik amacıyla)

		// İmza ve PublicKey'i parçalara ayırır
		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)]) // İmzanın ilk yarısı r değeri olarak ayarlanır
		s.SetBytes(in.Signature[(sigLen / 2):]) // İmzanın ikinci yarısı s değeri olarak ayarlanır

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)]) // PublicKey'in ilk yarısı x değeri olarak ayarlanır
		y.SetBytes(in.PubKey[(keyLen / 2):]) // PublicKey'in ikinci yarısı y değeri olarak ayarlanır

		rawPubKey := ecdsa.PublicKey{curve, &x, &y} // Raw public key oluşturulur
		// ECDSA algoritması kullanarak imza doğrulaması yapılır
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false // İmza doğrulanamazsa false döner
		}
	}

	return true // İşlem geçerli ise true döner
}

// TrimmedCopy fonksiyonu, Transaction yapısının bir kopyasını oluşturur ve girdileri ve çıktıları temizler.
// Temizlenmiş kopya, işlemi imzalamak veya doğrulamak için kullanılırken orijinal Transaction yapısını değiştirmez.
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput   // Boş bir TxInput (girdi) dizisi oluşturulur
	var outputs []TxOutput // Boş bir TxOutput (çıktı) dizisi oluşturulur

	// Orijinal işlemin girdilerini temizlenmiş kopyaya ekler
	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil}) // Girdinin sadece ID ve Out değerlerini kopyaya ekler
	}

	// Orijinal işlemin çıktılarını temizlenmiş kopyaya ekler
	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PublicKey}) // Çıktının sadece Value ve PublicKey değerlerini kopyaya ekler
	}

	// Temizlenmiş kopya Transaction yapısını oluşturur
	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy // Oluşturulan temizlenmiş kopyayı döndürür
}

// String fonksiyonu, Transaction yapısını stringe dönüştürür
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("\033[35m ╔═══════════════════════════════════════════════════════════════════════════════════"))
	lines = append(lines, fmt.Sprintf("\033[97m║\033[35m  ║ --- Transaction %x:\033[0m", tx.ID))

	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("\033[97m║\033[38;5;94m  ║   Input %d:\033[0m", i))

		lines = append(lines, fmt.Sprintf("\033[97m║\033[33m  ║     TXID:     %x\033[0m", input.ID))
		lines = append(lines, fmt.Sprintf("\033[97m║\033[33m  ║     Out:      %d\033[0m", input.Out))
		lines = append(lines, fmt.Sprintf("\033[97m║\033[33m  ║     Signature:%x\033[0m", input.Signature))
		lines = append(lines, fmt.Sprintf("\033[97m║\033[33m  ║     PubKey:   %x\033[0m", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("\033[97m║\033[34m  ║   Output %d:\033[0m", i))
		lines = append(lines, fmt.Sprintf("\033[97m║\033[36m  ║     Value:  %d\033[0m", output.Value))
		lines = append(lines, fmt.Sprintf("\033[97m║\033[36m  ║     Script: %x\033[0m", output.PublicKey))
	}

	lines = append(lines, fmt.Sprintf("\033[97m║\033[35m  ╚═══════════════════════════════════════════════════════════════════════════════════\033[0m"))

	return strings.Join(lines, "\n")
}
