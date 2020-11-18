package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/incognitokey"
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
				var keySet incognitokey.KeySet
				txList[0].Proof.GetInputCoins()[0].CoinDetails.GetSerialNumber()
				coin := txList[0].Proof.GetOutputCoins()[0]
				if coin.CoinDetailsEncrypted != nil && !coin.CoinDetailsEncrypted.IsNil() {
					if len(keySet.ReadonlyKey.Rk) > 0 {
						// try to decrypt to get more data
						err := coin.Decrypt(keySet.ReadonlyKey)
						if err != nil {
							panic(err)
						}
					}
				}
				coin.CoinDetails.GetSerialNumber()
				fmt.Println(shardID, len(txList))
			}
		}(shardInfo.ShardID, shardInfo.Height)
	}
	select {}
}
