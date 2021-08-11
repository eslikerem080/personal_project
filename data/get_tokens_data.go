package main

import(
	token_data "./token_data"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"os"
	"bufio"
	"strings"
	"fmt"
    //"sync"
    "math/big"
    "context"
)

func main() {
	
    client, err := ethclient.Dial("http://localhost:8889")
    
    //"https://mainnet.infura.io/v3/7ee001f2b684469faff12e0485f3f977"
    if err != nil {
        log.Fatal(err)
    }

	// Reading token addresses from the file

	file, err := os.Open("usdt_address.txt")
  
    if err != nil {
        log.Fatalf("failed to open")
  
    }
  
    scanner := bufio.NewScanner(file)
  
    scanner.Split(bufio.ScanLines)
    var text []string
  
    for scanner.Scan() {
        text = append(text, scanner.Text())
    }
  
    file.Close()
	var addresses [][]string
    for _, each_ln := range text {
		words := strings.Fields(each_ln)
        addresses = append(addresses, words)
    }

	fmt.Println(addresses)

    //var wg sync.WaitGroup

    fmt.Println(len(addresses))
    //wg.Add(360)

    // getting latest block

    header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
    end_block := header.Number
    start_block := big.NewInt(0).Add(end_block, big.NewInt(-36000))
    print(end_block)
    print(start_block)
    for _, address := range addresses {
        for i := start_block; i.Cmp(end_block) == -1 ; i = i.Add(i, big.NewInt(100)){
            last_block := big.NewInt(0)
            last_block = last_block.Add(i, big.NewInt(99))
            fmt.Println(i, "i")
            fmt.Println(last_block, "last_block")
            //go token_data.Get_token_data(address[1], address[0], client, &wg, i, last_block)
            token_data.Get_token_data(address[1], address[0], client, /*&wg,*/ i, last_block)
        }
        //wg.Wait()
        fmt.Println("All goroutines finished")
    }
    
    
}