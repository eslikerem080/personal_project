
package token_data

import (
    "fmt"
    "log"
	"encoding/csv"
	"os"

    //"math"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
	"context"
	"strconv"
	"time"

	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"sync"
     usdt "./token"
)

type Edge struct {

	address_from string
	address_to string
	value   float64
}

type key struct {
    address_from string
    address_to  string
}


var (
	_ = big.NewInt
)

func Get_token_data(address string, token_name string, client *ethclient.Client, wg *sync.WaitGroup) {

	defer wg.Done()

    // AAVE data contract address
    tokenAddress := common.HexToAddress(address)
	
    instance, err := usdt.NewToken(tokenAddress, client)

	if err != nil {
        log.Fatal(err)
    }

	decimal_places, err := instance.Decimals(&bind.CallOpts{})

    if err != nil {
        log.Fatal(err)
    }

	oldest_block, latest_block := getOldestBlock(client, 5)
	edges, dict_of_edges := getTransactionsData(tokenAddress, oldest_block, latest_block, client, decimal_places)
	
	var rows [][]string
	var nodes []string
	for i := 0; i < len(edges); i++ {
		fmt.Println("prints edges")
		fmt.Println("Edge ", i)
		fmt.Println(edges[i].address_from)
		fmt.Println(edges[i].address_to)
		fmt.Println(edges[i].value)
		v := strconv.FormatFloat(edges[i].value, 'f', 7, 64)
		data_to_add := []string{edges[i].address_from, edges[i].address_to, v}
		rows = append(rows, data_to_add)
		nodes = append(nodes, edges[i].address_from)
		nodes = append(nodes, edges[i].address_to)

	}
	name := "all_edges_" + token_name + ".csv"
	csvfile, err := os.Create(name)
 
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
 
	csvwriter := csv.NewWriter(csvfile)
	heading := []string{"idFrom", "idTo", "value"}
	_ = csvwriter.Write(heading)
	for _, row := range rows {
		_ = csvwriter.Write(row)
	}
 
	csvwriter.Flush()
 
	csvfile.Close()

	nodes = removeDuplicateValues(nodes)

	name = "nodes_" + token_name + ".csv"
	csvfile, err = os.Create(name)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
 
	csvwriter = csv.NewWriter(csvfile)

	heading = []string{"id"}
	_ = csvwriter.Write(heading)

	for _, node := range nodes {
		array_node := []string{node}
		_ = csvwriter.Write(array_node)
	}
 
	csvwriter.Flush()
 
	csvfile.Close()

	// appending data from dict 

	name = "aggregated_edges_" + token_name + ".csv"
	csvfile, err = os.Create(name)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter = csv.NewWriter(csvfile)
	

	heading = []string{"addressFrom", "addressTo", "value"}
	_ = csvwriter.Write(heading)

	
	for k, v := range dict_of_edges { 
		value := strconv.FormatFloat(v, 'f', 7, 64)
		data_to_add := []string{k.address_from, k.address_to, value}	
		_ = csvwriter.Write(data_to_add)
	}
 
	csvwriter.Flush()
 
	csvfile.Close()


}

func getOldestBlock(client *ethclient.Client, daysAgo int) (*big.Int, *big.Int) {

	var current_block *big.Int
	var oldest_block *big.Int
	current_block = big.NewInt(0)

	// Get current block
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	current_block = header.Number

	//2)  Find oldest block in our lookup date range
	oldest_block = new(big.Int).Set(current_block)

	now := time.Now()

	//timeonehourago := uint64(now.Add(-2*time.Hour).Unix())
	//timeonemonthago := uint64((now.AddDate(0, 0, -1)).Unix())
	timeonemonthago := uint64(now.Unix()) - 24*60*60*uint64(daysAgo)

	var j int64
	j = 0

	for {
		j -= 500
		oldest_block.Add(oldest_block, big.NewInt(j))

		block, err := client.BlockByNumber(context.Background(), oldest_block)
		if err != nil {
			log.Fatal(err)
		}

		if block.Time() < timeonemonthago {

			break
		}
	}

	return oldest_block, current_block
}

func decodeBytes(log *types.Log) *big.Int {
	
	if(len(log.Data) >= 31){
		value := new(big.Int).SetBytes(log.Data[0:32])
		return value
	}

	return big.NewInt(-1)

}

func getTransferFromTxLog(logs []*types.Log, pooltopics []string) (string, string, *big.Int) {
	//fmt.Println("getTransferFromTxLog")
	var firstLog *types.Log
	var address_from string
	var address_to string
	//var lastLog *types.Log

	for _, log := range logs {
		
		if len(log.Topics) > 0 {
			if log.Topics[0] != common.HexToHash(pooltopics[0]) { // Index out of range fix
				continue
			}
			if firstLog == nil {

				firstLog = log
				address_from = HashToReserveAddress(log.Topics[1])
				address_to = HashToReserveAddress(log.Topics[2])
			}
			//lastLog = log
		}
	}

	

	if firstLog == nil { // could not find any valid swaps, thus the transaction failed
		return "none", "none", common.Big0
	}
	value := decodeBytes(firstLog)

	return address_from, address_to, value
}

// Generates Transaction Logs
func getTransactionsData(pool_address common.Address, oldest_block *big.Int, latest_block *big.Int, client *ethclient.Client, decimals uint8) ([]Edge, map[key] float64) {
	
	var edges []Edge
	var m = make(map[key] float64)
	var logsX []types.Log

	poolTopics := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

	big_ten := big.NewInt(1000)
	number_of_blocks := big.NewInt(0).Sub(latest_block, oldest_block)
	step := big.NewInt(0).Div(number_of_blocks, big_ten)
	fmt.Println(step)
	fmt.Println(number_of_blocks)
	new_oldest_block := oldest_block

	for i:= big.NewInt(0).Add(oldest_block, step); i.Cmp(latest_block) == -1; i.Add(i, step){
	
		logs, err := try_new_query(pool_address, new_oldest_block, i, client)
		if err != nil {
			log.Fatal(err)
		}
		logsX = append(logsX, logs...)
		new_oldest_block = i
		new_oldest_block = big.NewInt(0).Add(new_oldest_block, big.NewInt(1))
	}

	


	// Filters transaction logs
	
		number_of_logs := len(logsX)
		//fmt.Println("looping through an array of logs")
		//fmt.Println(number_of_logs)
		for i := 0; i < number_of_logs; i++ {
			
			if logsX[i].Topics[0] != common.HexToHash(poolTopics[0]) {
				continue
			}

			txlog, err := client.TransactionReceipt(context.Background(), logsX[i].TxHash)

			if err != nil {
				log.Fatal(err)
			}

			// add to edges
			address_from, address_to, value := getTransferFromTxLog(txlog.Logs, poolTopics)
			v := add_decimals(value, decimals)
			value_string := v.String()
			v_float64, err := strconv.ParseFloat(value_string, 64)

			if err != nil {
				log.Fatal(err)
			}
			if val, ok := m[key{address_from: address_from, address_to: address_to}]; ok {
				m[key{address_from: address_from, address_to: address_to}] = val + v_float64
			}else{
				m[key{address_from: address_from, address_to: address_to}] = v_float64
			}
			
			edges = append(edges, Edge{address_from, address_to, v_float64})

			
		}
	
	return edges, m
}

func HashToReserveAddress(hash common.Hash) string {
	var value []string
	value = append(value, "0", "x")
	value = append(value, hex.EncodeToString(hash[12:32]))
	valueStr := strings.Join(value, "")

	return valueStr
}

func removeDuplicateValues(intSlice []string) []string {
    keys := make(map[string]bool)
    list := []string{}
  
    // If the key(values of the slice) is not equal
    // to the already present value in new slice (list)
    // then we append it. else we jump on another element.
    for _, entry := range intSlice {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}

func add_decimals(value *big.Int, decimals uint8) *big.Float {
	f := new(big.Float).SetInt(value)
	float_decimals := float64(decimals)
	d :=  big.NewFloat(float_decimals)
	for i := 0; i < int(decimals); i++{
		f = new(big.Float).Quo(f, d)
	}

	return f
}

func try_new_query(pool_address common.Address, oldest_block *big.Int, latest_block *big.Int, client *ethclient.Client) ([]types.Log, error){
	
	new_oldest_block := oldest_block
	query := ethereum.FilterQuery{

		FromBlock: new_oldest_block,
		ToBlock:   latest_block, // = latest block
		Addresses: []common.Address{pool_address},
		
	}

	var err error

	logsX, err := client.FilterLogs(context.Background(), query)
	if(err != nil){
		fmt.Println("enters recursion")
		new_oldest_block = big.NewInt(0).Add(oldest_block, big.NewInt(100))
		if(new_oldest_block.Cmp(latest_block) != -1){
			logsX, err = try_new_query(pool_address, new_oldest_block, latest_block, client)
		}
	}

	err = nil
	return logsX, err

}