[![Go](https://github.com/SadikSunbul/GO-BlockChain-Simulation/actions/workflows/go.yml/badge.svg)](https://github.com/SadikSunbul/GO-BlockChain-Simulation/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SadikSunbul/GO-BlockChain-Simulation)](https://goreportcard.com/report/github.com/SadikSunbul/GO-BlockChain-Simulation)
[![Version](https://img.shields.io/badge/Version-1.0-blue)]()

# GO-BlockChain-Simulation

GO-BlockChain-Simulation is a blockchain simulation written in Go. This project is developed to understand and implement the fundamentals of blockchain technology.
***

<img src="https://repository-images.githubusercontent.com/815066994/a67ad049-f2ce-4276-8f24-5f01ceefaf30">

## Project Structure

The project consists of the following components:

- **cli**: Contains necessary code for the command-line interface.
- **blockchain**: Defines the structure and operations of the blockchain.
- **wallet**: Provides operations to manage cryptocurrency wallets.
- **main.go**: Main application file of the project.
***

## Installation

Clone the project to your local machine:

+ ```bash
   $ git clone https://github.com/SadikSunbul/GO-BlockChain-Simulation.git
   $ cd GO-BlockChain-Simulation 
***

## Usage

### Creating a Blockchain

To create a new blockchain:

+ ```bash 
   $ go run main.go createblockchain -address <ADDRESS> 
***

### Checking Balance

To check the balance for a specific address:

+ ```bash
   $ go run main.go getbalance -address <ADDRESS>
***

### Sending Transactions

To send a transaction on the blockchain:

+ ```bash
   $ go run main.go send -from <FROM_ADDRESS> -to <TO_ADDRESS> -amount <AMOUNT> -mine
***

### Viewing the Blockchain

To print all blocks in the blockchain:

+ ```bash
   $ go run main.go printchain 
***

### Creating a New Wallet

To create a new wallet:

+  ```bash
   $ go run main.go createwallet
***

### List Wallets

Lists the public keys of the wallets you created on your device:

+ ```bash
   $ go run main.go listaddresses
***

### Reindex UTXO

+ ```bash
   $ go run main.go reindexutxo
***

### Start Node

+ ```bash
   $ go run main.go startnode -miner <ADDRESS>
***

## Contributing

If you would like to contribute, please open a pull request on [GitHub](https://github.com/SadikSunbul/GO-BlockChain-Simulation). We welcome contributions of any kind to the project.
This project follows a [Code of Conduct](CODE_OF_CONDUCT.md). Please review and adhere to it in all interactions within the project.
***
