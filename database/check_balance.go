package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"

	tb "github.com/tigerbeetle/tigerbeetle-go"
)

func accountIDFromString(accountID string) uint64 {
	hash := sha256.Sum256([]byte(accountID))
	return binary.BigEndian.Uint64(hash[:8])
}

func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

func main() {
	addr := os.Getenv("TIGERBEETLE_ADDR")
	if addr == "" {
		addr = "172.28.0.10:3000"
	}

	client, err := tb.NewClient(tb.ToUint128(0), []string{addr})
	if err != nil {
		fmt.Println("Error conectando:", err)
		return
	}
	defer client.Close()

	accountStrings := []string{"4001-404c-e02d-a614", "4001-f308-ce29-4fbb"}

	for _, accStr := range accountStrings {
		tbID := accountIDFromString(accStr)
		ids := []tb.Uint128{tb.ToUint128(tbID)}

		accounts, err := client.LookupAccounts(ids)
		if err != nil {
			fmt.Printf("%s: error %v\n", accStr, err)
			continue
		}
		if len(accounts) == 0 {
			fmt.Printf("%s: CUENTA NO ENCONTRADA en TigerBeetle (tbID=%d)\n", accStr, tbID)
			continue
		}

		creditsBytes := accounts[0].CreditsPosted.Bytes()
		debitsBytes := accounts[0].DebitsPosted.Bytes()
		reverseBytes(creditsBytes[:])
		reverseBytes(debitsBytes[:])

		credits := new(big.Int).SetBytes(creditsBytes[:])
		debits := new(big.Int).SetBytes(debitsBytes[:])
		balance := new(big.Int).Sub(credits, debits)

		fmt.Printf("%s (tbID=%d):\n", accStr, tbID)
		fmt.Printf("  CreditsPosted: %s centavos\n", credits.String())
		fmt.Printf("  DebitsPosted:  %s centavos\n", debits.String())
		fmt.Printf("  Balance:       %s centavos ($%.2f)\n\n", balance.String(), float64(balance.Int64())/100)
	}
}