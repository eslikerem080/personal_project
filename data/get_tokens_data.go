package main

import(
	token_data "./token_data"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"os"
	"bufio"
	"strings"
	"fmt"
    "sync"
)

func main() {
	
    client, err := ethclient.Dial("http://localhost:8888")
    //"http://localhost:8888"
    //"https://mainnet.infura.io/v3/7ee001f2b684469faff12e0485f3f977"
    if err != nil {
        log.Fatal(err)
    }

	// Reading token addresses from the file

	file, err := os.Open("token_addresses.txt")
  
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

    var wg sync.WaitGroup

    fmt.Println(len(addresses))
    wg.Add(len(addresses))
    for _, address := range addresses {
        go token_data.Get_token_data(address[1], address[0], client, &wg)
    }
    wg.Wait()
    fmt.Println("All goroutines finished")
}