package wallet

import (
	"fmt"
	"testing"

	"github.com/Ulugbek999/wallet/pkg/types"

	"github.com/google/uuid"
)

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}

	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposti account, error = %v", err)
	}

	return account, nil
}

func TestService_FindAccountByID_found(t *testing.T) {
	svc := newTestService()
	account, err := svc.addAccountWithBalance("+992001010000", 1)
	if err != nil {
		t.Error(err)
	}
	_, err = svc.FindAccountByID(account.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestService_FindAccountByID_notFound(t *testing.T) {
	svc := &Service{}
	_, err := svc.RegisterAccount("+992001010000")
	if err != nil {
		t.Error(err)
	}
	var accountID int64 = 123
	_, err = svc.FindAccountByID(accountID)
	if err == nil {
		t.Errorf("User with ID %v found", accountID)
	}
}

func TestService_FindPaymentByID_found(t *testing.T) {
	svc := &Service{}
	account, err := svc.RegisterAccount("+992001010000")
	svc.Deposit(account.ID, 20_000_00)
	if err != nil {
		t.Error(err)
	}
	payment, err := svc.Pay(account.ID, 10_000_00, "Mobile")
	if err != nil {
		t.Error(err)
	}
	_, err = svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestService_FindPaymentByID_notFound(t *testing.T) {
	svc := &Service{}
	paymentID := uuid.New().String()
	_, err := svc.FindPaymentByID(paymentID)
	if err == nil {
		t.Errorf("Found payment by ID: %v", paymentID)
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := newTestService()

	phone := types.Phone("+992001010000")
	account, err := svc.addAccountWithBalance(phone, 10_000_00)
	if err != nil {
		t.Error(err)
		return
	}

	payment, err := svc.Pay(account.ID, 1000_0, "auto")
	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}

	err = svc.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}
}

func TestService_Reject_paymentFound(t *testing.T) {
	svc := &Service{}
	account, err := svc.RegisterAccount("+992001010000")
	svc.Deposit(account.ID, 20_000_00)
	if err != nil {
		t.Error(err)
	}
	payment, err := svc.Pay(account.ID, 10_000_00, "Mobile")
	if err != nil {
		t.Error(err)
	}
	err = svc.Reject(payment.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestService_Reject_paymentNotFound(t *testing.T) {
	svc := &Service{}
	paymentID := uuid.New().String()
	err := svc.Reject(paymentID)
	if err == nil {
		t.Errorf("Found payment by ID: %v", paymentID)
	}
}

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()

	account, err := s.addAccountWithBalance("9127660305", 100)
	if err != nil {
		t.Errorf("account => %v", account)
		return
	}

	payment, err := s.Pay(account.ID, 10, "some")
	if err != nil {
		t.Errorf("payment => %v", payment)
		return
	}

	newPayment, err := s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("newPayment => %v", newPayment)
		return
	}
}