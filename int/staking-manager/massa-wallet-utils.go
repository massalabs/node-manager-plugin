package stakingManager

import (
	"fmt"

	"github.com/awnumar/memguard"
	WalletPkg "github.com/massalabs/station-massa-wallet/pkg/wallet"
)

// getPrivateKeyFromNickname returns the private key and address from a nickname
func getPrivateKeyFromNickname(pwd, nickname string) (string, string, error) {
	wallet, err := WalletPkg.New("")
	if err != nil {
		return "", "", fmt.Errorf("failed to create wallet from nickname %s: %v", nickname, err)
	}

	account, err := wallet.Load(nickname)
	if err != nil {
		return "", "", fmt.Errorf("failed to load account from nickname %s: %v", nickname, err)
	}

	address, err := account.Address.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to get address from nickname %s: %v", nickname, err)
	}

	privateKey, err := account.PrivateKeyTextInClear(memguard.NewBufferFromBytes([]byte(pwd)))
	if err != nil {
		return "", "", fmt.Errorf("failed to get private key from nickname %s: %v", nickname, err)
	}

	return privateKey.String(), address, nil
}
