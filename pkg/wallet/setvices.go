package wallet

import (
	"errors"

	"github.com/Ulugbek999/wallet/pkg/types"
)

var (
	ErrPhoneRegistered = errors.New("phone already registered")
	ErrAccountNotFound = errors.New("account not found")
)

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	// payments      []*types.Payment
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}






