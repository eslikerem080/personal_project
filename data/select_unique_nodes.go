package main 

import(
	"os"
	"encoding/csv"
	"log"

)

func main(){

	uni, err := os.Open("nodes_uni.csv")
    if err != nil {
        
    }
    defer uni.Close() // this needs to be after the err check

    lines_uni, err := csv.NewReader(uni).ReadAll()
    if err != nil {
        log.Fatal(err)
    }

	dai, err := os.Open("nodes_dai.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer dai.Close() // this needs to be after the err check

    lines_dai, err := csv.NewReader(dai).ReadAll()
    if err != nil {
        log.Fatal(err)
    }


	all_nodes := append(convert_2d_to_1d(lines_uni), convert_2d_to_1d(lines_dai)...)

	all_nodes = removeDuplicateValues(all_nodes)

	name := "all_nodes.csv"
	csvfile, err := os.Create(name)
 
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
 
	csvwriter := csv.NewWriter(csvfile)
	heading := []string{"id"}
	_ = csvwriter.Write(heading)
	
	for _, row := range all_nodes {
		array_row := []string{row}
		_ = csvwriter.Write(array_row)
	}
 
	csvwriter.Flush()
 
	csvfile.Close()

	
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

func convert_2d_to_1d(array [][]string) []string {
	var new_array []string

	for _, line := range(array) {
		if(line[0] != "id"){
			new_array = append(new_array, line[0])
		}
	}

	return new_array
}