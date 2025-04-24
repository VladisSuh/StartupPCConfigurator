package main

import (
	"StartupPCConfigurator/internal/aggregator/parser/citilink"
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "[DNS Test] ", log.LstdFlags)
	parser := citilink.NewCitilinkParser(logger)

	url := "https://www.citilink.ru/product/videokarta-palit-pci-e-4-0-rtx4060-infinity-2-nv-rtx4060-8gb-128bit-gd-2010430/"

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// <-- CALL THE LOW‑LEVEL PARSER DIRECTLY -->
	prod, err := parser.ParseProductPage(ctx, url)
	if err != nil {
		logger.Fatalf("ParseProductPage error: %v", err)
	}

	// NOW prod *has* Name, Price, Description, Availability, etc.
	fmt.Printf("Name:            %s\n", prod.Name)
	fmt.Printf("Price:           %s\n", prod.Price)
	fmt.Printf("Availability:    %s\n", prod.Availability)
	fmt.Printf("Category:        %s\n", prod.Category)
	fmt.Printf("Main Image:      %s\n", prod.MainImage)
	fmt.Printf("Description:     %s\n", prod.Description)
	for i, kv := range prod.Characteristics {
		fmt.Printf("Spec %d: %s = %s\n", i+1, kv.Key, kv.Value)
	}
}
