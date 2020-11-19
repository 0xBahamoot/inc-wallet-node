package database

import "sync"

// shard based
type CoinDatabase struct {
	lock        sync.RWMutex
	TokenShards map[string]TokenCoinDatabase //key: shardID+tokenID
}

type TokenCoinDatabase struct {
}

func (db *CoinDatabase) SaveCoin(coinData []byte, hash []byte, shardID byte, tokenID []byte) error {
	return nil
}

func (db *CoinDatabase) GetAllCoins(shardID byte, tokenID []byte) ([][]byte, error) {
	return nil, nil
}

func (db *CoinDatabase) GetAllUserCoins(shardID byte, tokenID []byte, key string) ([][]byte, error) {
	return nil, nil
}

func (db *CoinDatabase) GetUTXOCoins(shardID byte, tokenID []byte, key string) ([][]byte, error) {
	return nil, nil
}
