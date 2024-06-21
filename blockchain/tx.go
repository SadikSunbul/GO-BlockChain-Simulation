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

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PublicKey = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PublicKey, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}
