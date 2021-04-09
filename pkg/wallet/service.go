package wallet

import (
	"errors"

	"github.com/Ulugbek999/wallet/pkg/types"
	"github.com/google/uuid"
)

//Error variables.
var (
	ErrPhoneRegistered      = errors.New("phone already registered")
	ErrAccountNotFound      = errors.New("account not found")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrNotEnoughBalance     = errors.New("balance is not anough")
	ErrPaymentNotFound      = errors.New("payment not found")
)

//Service - service struct.
type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
}

//RegisterAccount - authentication processes method performing.
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

// Deposit -  replenish the user's account.
func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

//Pay - payments method.
func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

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

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil

}

//FindAccountByID - method that find account by ID.
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {

	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}
	return nil, ErrAccountNotFound
}

//FindPaymentByID - method that find payment by ID.
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {

	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

//Reject - method that returns payment in a accident of error.
func (s *Service) Reject(paymentID string) error {

	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return ErrPaymentNotFound
	}

	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount

	return nil
}






