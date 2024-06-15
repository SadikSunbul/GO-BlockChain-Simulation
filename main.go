package main

import (
	"fmt"
	"github.com/SadikSunbul/GO-BlockChain-Simulation/blockchain"
	"strconv"
)

func main() {

	chain := blockchain.InitBlockChain() //zinciri başlattım

	chain.AddBlock("1. Blok bir öncesi genesis") //zincire blok ekledik
	chain.AddBlock("2. Blok bir öncesi 1. blok")
	chain.AddBlock("3. Blok bir öncesi 2. blok")

	//zinciri okuycaz

	for _, block := range chain.Block {
		fmt.Printf("Önceki Blog Hasi :%x\n", block.PrevHash)
		fmt.Printf("Blog Verisi      :%s\n", block.Data)
		fmt.Printf("Blog Hasi        :%x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("Pow:%s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
