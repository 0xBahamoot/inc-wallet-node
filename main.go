package main

import (
	"fmt"
	"time"
)

//modes:
// 1. Public
// 2. Private (use readonly key)
func main() {
	endpoint := "http://51.83.237.20:9338"
	shardInfos, err := GetBlockChainInfo(endpoint)
	if err != nil {
		panic(err)
	}
	fmt.Println(shardInfos)
	t := time.Now().Unix()
	for _, shardInfo := range shardInfos {
		go func(shardID byte, height uint64) {
			coinCount := 0
			for i := uint64(1); i <= 200; i++ {
				txList, err := RetrieveShardBlockTxs(endpoint, shardID, i)
				if err != nil {
					panic(err)
				}
				for _, tx := range txList {
					// for _, coin := range tx.Proof.GetInputCoins() {
					// 	coin.CoinDetails.GetSerialNumber()
					// 	// coin.CoinDetails.GetSerialNumber().ToBytesS()
					// }
					for _, _ = range tx.Proof.GetOutputCoins() {
						// coin.CoinDetails.HashH().String()
						coinCount++
					}
				}
			}
			fmt.Println(shardID, coinCount)
			fmt.Println(time.Now().Unix() - t)
		}(shardInfo.ShardID, shardInfo.Height)
	}
	select {}
}

// save usedCoinSN -> delete usedCoin
// save OutputCoin

// if coin.CoinDetailsEncrypted != nil && !coin.CoinDetailsEncrypted.IsNil() {
// 	if len(keySet.ReadonlyKey.Rk) > 0 {
// 		// try to decrypt to get more data
// 		err := coin.Decrypt(keySet.ReadonlyKey)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
