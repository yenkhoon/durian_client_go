package main

import "github.com/durian-client-go/durian"

// hashServer is a local implementation of Hash.
type Provider struct {
}

func (provider Provider) Exist(durian.Provider_exist) error { return nil }
func (provider Provider) Account(p durian.Provider_account) error {
	acc, err := p.Results.NewAccount()
	if err != nil {
		return nil
	}
	acc.SetBalance([]byte("abcdeabcdeabcdeabcde"))
	acc.SetNonce([]byte{0})
	acc.SetCode(nil)

	return nil
}
func (provider Provider) UpdateAccount(durian.Provider_updateAccount) error   { return nil }
func (provider Provider) CreateContract(durian.Provider_createContract) error { return nil }
func (provider Provider) StorageAt(durian.Provider_storageAt) error           { return nil }
func (provider Provider) SetStorage(durian.Provider_setStorage) error         { return nil }
func (provider Provider) Timestamp(durian.Provider_timestamp) error           { return nil }
func (provider Provider) BlockNumber(durian.Provider_blockNumber) error       { return nil }
func (provider Provider) BlockHash(durian.Provider_blockHash) error           { return nil }
func (provider Provider) BlockAuthor(durian.Provider_blockAuthor) error       { return nil }
func (provider Provider) Difficulty(durian.Provider_difficulty) error         { return nil }
func (provider Provider) GasLimit(durian.Provider_gasLimit) error             { return nil }
