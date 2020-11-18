package database

type UsedSerialNumber struct {
	ShardID byte
	TokenID []byte
}

func (db *UsedSerialNumber) SaveUsedSerialNumber(coinHash []byte, publickey []byte) error {
	return nil
}

func (db *UsedSerialNumber) CheckUsedSerialNumber(hash []byte) error {
	return nil
}
