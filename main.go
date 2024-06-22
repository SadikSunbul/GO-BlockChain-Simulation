package main

import (
	"github.com/SadikSunbul/GO-BlockChain-Simulation/cli"
	"os"
)

func main() {
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()
}
