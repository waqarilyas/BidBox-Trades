package utils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func initCoinPiars() ([]string, error) {

	var coin_pairs []string

	// Open the CSV file
	file, err := os.Open("test.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}, err
	}

	// Iterate over the records and print the values
	for _, record := range records {
		for i, value := range record {
			if i == 1 {
				coin_pairs = append(coin_pairs, value)
			}
		}
	}
	return coin_pairs, nil
}

func AiStub(orders int) ([]string, []string, error) {
	rand.Seed(time.Now().UnixNano()) // initialize the random number generator with the current time

	// define an array of integers
	coins := []string{"SETHSUSDT_SUMCBL"}
	// create a slice to hold the selected elements
	list := make([]string, orders)

	// generate the random indices and select the corresponding elements
	for i := 0; i < orders; i++ {
		index := rand.Intn(len(coins)) // generate a random index within the range of the array
		list[i] = coins[index]         // select the element at the random index and add it to the selected slice
	}

	middle := orders / 2
	return list[middle:], list[:middle], nil
}

func CheckBalance(val float64) error {
	if val < 200 {
		return errors.New("Balance should be at least 200 USDT")
	}
	return nil
}
