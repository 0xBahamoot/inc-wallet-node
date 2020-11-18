package database

import "sync"

// shard based
type CoinDatabase struct {
	lock        sync.Mutex
	TokenShards map[string]TokenCoinDatabase //key: shardID+tokenID
}

type TokenCoinDatabase struct {
	lock sync.RWMutex
}

func (db *CoinDatabase) SaveCoin(coinData []byte, hash []byte, shardID byte, tokenID []byte) error {
	return nil
}

func (db *CoinDatabase) GetCoins(shardID byte, tokenID []byte) ([][]byte, error) {
	return nil, nil
}
