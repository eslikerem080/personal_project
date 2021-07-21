package main

import(
	"fmt"
	uni "./uniswap_factory"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	//"os"
	//"bufio"
	//"strings"
    //"sync"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func main(){

	client, err := ethclient.Dial("https://mainnet.infura.io/v3/e009cbb4a2bd4c28a3174ac7884f4b42")
	if err != nil {
		log.Fatal(err)
	}

	factory_address := common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")

	uniswap_factory, err := uni.NewUniswapFactory(factory_address, client)

	if err != nil {
		log.Fatal(err)
	}

	num_addresses, err := uniswap_factory.AllPairsLength(&bind.CallOpts{})

	if err != nil {
		log.Fatal(err)
	}
	
	for i:= big.NewInt(0); i.Cmp(num_addresses) == -1; i.Add(i, big.NewInt(1)) {
		fmt.Println(i)
		fmt.Println(uniswap_factory.AllPairs(&bind.CallOpts{}, i))
	}

}