package main

import (
	"fmt"
)

func main() {
	endpoint := "http://51.83.237.20:9338"
	shardInfos, err := GetBlockChainInfo(endpoint)
	if err != nil {
		panic(err)
	}
	fmt.Println(shardInfos)
	for _, shardInfo := range shardInfos {
		go func(shardID byte, height uint64) {
			for i := uint64(1); i < 10; i++ {
				txList, err := RetrieveShardBlockTxs(endpoint, shardID, i)
				if err != nil {
					panic(err)
				}
				for _, tx := range txList {
					for _, coin := range tx.Proof.GetInputCoins() {
						coin.CoinDetails.GetSerialNumber()
						// coin.CoinDetails.GetSerialNumber().ToBytesS()
					}
					for _, coin := range tx.Proof.GetOutputCoins() {
						coin.CoinDetails.HashH().Bytes()
					}
				}

				fmt.Println(shardID, len(txList))
			}
		}(shardInfo.ShardID, shardInfo.Height)
	}
	select {}
}

// save usedCoinHash -> delete usedCoin
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
