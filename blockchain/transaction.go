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
	"github.com/SadikSunbul/GO-BlockChain-Simulation/wallet"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	ID      []byte     //transectıon hası
	Inputs  []TxInput  //bu transectıondakı ınputlar
	Outputs []TxOutput //bu transectıondakı outputlar
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" { //data boş ise gir
		data = fmt.Sprintf("Coins to %s", to) //paralar to da der
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)} //hıcbır cıktıya referabs vermez ,cıkıs endexi -1 aynı referans yok , sadce data mesajı vardır
	txout := NewTXOutput(100, to)                    //100 tokeni to ya gonderırı

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}} //transectıonı olustururuz
	tx.SetID()                                                  //Transectıon Id sını olustururuz
	return &tx
}

func (tx *Transaction) SetID() { //Id olusturur transectıonun
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx) //transectıonu encode edıyoruz
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes()) //transectıonu byte seklınde sha256 ıle sıfrelıyoruz ve ıd yı urettık
	tx.ID = hash[:]
}

// NewTransaction, belirtilen bir adresten başka bir adrese belirtilen miktar token transferi yapacak yeni bir işlem oluşturur.
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)

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

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
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

func (tx Transaction) Serilize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)

	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serilize())
	return hash[:]
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PublicKey
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature

	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PublicKey
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PublicKey})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

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
