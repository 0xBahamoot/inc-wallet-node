package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/incognitochain/incognito-chain/transaction"
)

func GetBlockChainInfo(endpoint string) ([]ShardInfo, error) {
	var result []ShardInfo
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblockchaininfo",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(endpoint, requestBody)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result GetBlockChainInfoResult
		Error  *ErrMsg
	}
	_ = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, errors.New(resp.Error.StackTrace)
	}
	for chainID, chainInfo := range resp.Result.BestBlocks {
		if chainID >= 0 {
			result = append(result, ShardInfo{
				ShardID: byte(chainID),
				Height:  chainInfo.Height,
			})
		}
	}
	return result, nil
}

func RetrieveShardBlockTxs(endpoint string, shardID byte, height uint64) ([]transaction.Tx, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "retrieveblockbyheight",
		"params":  []interface{}{float64(height), float64(shardID), "2"},
		"id":      1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(endpoint, requestBody)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []GetShardBlockResult
		Error  *ErrMsg
	}
	_ = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, errors.New(resp.Error.StackTrace)
	}
	var result []transaction.Tx
	for _, txHex := range resp.Result[0].Txs {
		txBytes, err := hex.DecodeString(txHex.HexData)
		if err != nil {
			return nil, err
		}
		var tx transaction.Tx
		err = json.Unmarshal(txBytes, &tx)
		if err != nil {
			return nil, err
		}
		result = append(result, tx)
	}
	return result, err
}

func sendRequest(endpoint string, requestBody []byte) ([]byte, error) {
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
