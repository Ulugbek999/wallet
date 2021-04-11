package wallet

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Ulugbek999/wallet/pkg/types"
)

type testService struct {
	*Service
}

func NewTestService() *testService {
	return &testService{Service: &Service{}}
}
func (s *testService) addAccount(phone types.Phone, balance types.Money) (*types.Account, *types.Payment, error) {
	if balance < 1 {
		return nil, nil, errors.New("Please give balance with positive number")
	}
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, nil, err
	}
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, nil, err
	}
	payment, err := s.Pay(account.ID, 1, "cafe")
	if err != nil {
		return nil, nil, err
	}
	return account, payment, nil
}
func TestRegisterAccount(t *testing.T) {
	wallet := NewTestService()
	_, _, err := wallet.addAccount("+99922292", 10)
	if err != nil {
		t.Error("for first user get Erro!!!(err)")
	}
	_, _, err = wallet.addAccount("+99922292", 10)
	if err == nil {
		t.Error("this user registered second times but we have not get Error!!!(err)")
	}
}
func TestFindAccountByID(t *testing.T) {
	var wallet Service
	_, err := wallet.RegisterAccount("+99999222222")
	if err != nil {
		t.Error("for first user get Erro!!!(err.2)")
	}
	_, err = wallet.FindAccountByID(1)
	if err != nil {
		t.Error("can not find Account by ID")
	}
}
func TestPay(t *testing.T) {
	wallet := NewTestService()
	account, _, err := wallet.addAccount("+123123", 10)
	if err != nil {
		t.Error("Pay(): for first user get Erro!!!(err(Pay))")
	}
	payment, err := wallet.Pay(1, types.Money(5), "cafe")
	if err != nil {
		t.Error("Pay(): something wrong while paying")
	}
	account1, _ := wallet.FindAccountByID(payment.AccountID)
	if account.Balance != account1.Balance {
		t.Error("Pay(): something wrong while paying without err")
	}
}
func TestReject(t *testing.T) {
	var wallet Service
	_, err := wallet.RegisterAccount("+99999222222")
	account, _ := wallet.FindAccountByID(1)
	account.Balance += 10
	if err != nil {
		t.Error("for first user get Error!!!(err.2)")
	}
	payment, err := wallet.Pay(1, types.Money(5), "cafe")
	if err != nil {
		t.Error("something wrong while paying")
	}
	if payment.Amount != 5 {
		t.Error("something wrong while paying without err")
	}
	err = wallet.Reject(payment.ID)
	account, _ = wallet.FindAccountByID(1)
	if account.Balance != 10 {
		t.Error("Reject is uncomplite!!")
	}
	if err != nil {
		t.Error(err)
	}
}
func TestFindPaymentByID(t *testing.T) {
	var wallet Service
	_, err := wallet.RegisterAccount("+99999222222")
	account, _ := wallet.FindAccountByID(1)
	account.Balance += 10
	if err != nil {
		t.Error("for first user get Error!!!(err.2)")
	}
	//succes
	payment, err := wallet.Pay(1, types.Money(5), "cafe")
	if err != nil {
		t.Error("something wrong while paying")
	}
	if payment.Amount != 5 {
		t.Error("something wrong while paying without err")
	}
	_, err1 := wallet.FindPaymentByID(payment.ID)
	if err1 != nil {
		t.Error("FinPaymentByID not correct")
	}
}
func TestDeposit(t *testing.T) {
	var wallet Service
	_, err := wallet.RegisterAccount("+99999222222")
	if err != nil {
		t.Error("for first user get Error!!!(err.2)")
	}
	err = wallet.Deposit(1, 10)
	if err != nil {
		t.Error("something wrong while paying")
	}
	account, _ := wallet.FindAccountByID(1)
	if account.Balance != 10 {
		t.Error("something wrong while paying without err")
	}
}
func TestRepeat(t *testing.T) {
	var init int = 100
	wallet := NewTestService()
	account, _, err := wallet.addAccount("+23123", types.Money(100))
	if err != nil {
		t.Errorf("Repeat(): Error(): can't add an account: %v", err)
	}
	payment, err := wallet.Pay(account.ID, types.Money(10), "cafe")
	if err != nil {
		t.Errorf("Repeat(): Error(): can't pay for an account(1): %v", err)
	}
	payment, err = wallet.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat(): Error(): Repeat not work: %v", err)
	}
	if types.Money((init-1)-2*10) != account.Balance {
		t.Error("Repeat(): Error(): something is wrong")
	}
}
func TestFavoritePayment(t *testing.T) {
	wallet := NewTestService()
	_, payment, err := wallet.addAccount("+212124", 100)
	if err != nil {
		t.Errorf("FavoritePayment(): Error: can't add account: %v", err)
	}
	favorite, err := wallet.FavoritePayment(payment.ID, "first")
	if err != nil {
		t.Errorf("FavoritePayment(): Error: FavoritePayment not work: %v", err)
	}
	expect := &types.Favorite{
		ID:        lastIDofFavorite,
		AccountID: payment.AccountID,
		Name:      "first",
		Amount:    payment.Amount,
		Category:  payment.Category,
	}
	if !reflect.DeepEqual(expect, favorite) {
		t.Errorf("FavoritePayment(): \nhave: %v\ngot: %v\n", expect, payment)
	}
}
func TestPayFromFavorite(t *testing.T) {
	wallet := NewTestService()
	_, payment, err := wallet.addAccount("+212124", 100) /// account.Balance = 50
	if err != nil {
		t.Errorf("PayFromFavorite(): Error: can't add account: %v", err)
	}
	favorite, err := wallet.FavoritePayment(payment.ID, "first")
	if err != nil {
		t.Errorf("PayFromFavorite(): Error: FavoritePayment not work: %v", err)
	}
	payment, err = wallet.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): Error: can't pay from favorite: %v", err)
	}
}