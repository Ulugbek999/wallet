package wallet

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestService_FindAccountByID_success(t *testing.T) {
	service := &Service{}
	service.RegisterAccount("+998204567898")

	account, err := service.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot: %v \nwant: nil  ", account)
	}
}

func TestService_FindAccountByID_notFound(t *testing.T) {
	service := Service{}
	service.RegisterAccount("+998204567898")

	_, err := service.FindAccountByID(2)
	if err == nil {
		t.Error(ErrAccountNotFound)
		return
	}
}

func TestService_FindPaymentByID_success(t *testing.T) {
	// create a service
	service := &Service{}

	// register user
	account, err := service.RegisterAccount("+998204567898")
	if err != nil {
		t.Errorf("Can not register an account, error: %v ", err)
		return
	}

	//  account top up
	err = service.Deposit(account.ID, 500)
	if err != nil {
		t.Errorf("Can not deposit an account, error: %v", err)
		return
	}

	// make a payment to account
	payment, err := service.Pay(account.ID, 100, "auto")
	if err != nil {
		t.Errorf("Can not create a payment, error: %v", err)
		return
	}

	// trying to cancel the payment
	result, err := service.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Payment not found, error: %v", err)
		return
	}

	// compare the payments
	if !reflect.DeepEqual(result, payment) {
		t.Errorf("Wrong payment, error: %v", err)
		return
	}
}

func TestService_FindPaymentByID_notFound(t *testing.T) {

	service := &Service{}

	// trying to find a non-existent payment
	paymentID := uuid.New().String()
	_, err := service.FindPaymentByID(paymentID)
	if err == nil {
		t.Errorf("Found payment by ID: %v", paymentID)
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("ErrPaymentNotFound: %v", err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	service := &Service{}

	account, err := service.RegisterAccount("+998204567898")
	if err != nil {
		t.Errorf("Can not register an account, error: %v ", err)
		return
	}

	err = service.Deposit(account.ID, 500)
	if err != nil {
		t.Errorf("Can not deposit an account, error: %v", err)
		return
	}

	payment, err := service.Pay(account.ID, 100, "cafe")
	if err != nil {
		t.Errorf("Can not create a payment, error: %v", err)
		return
	}

	// trying to cancel the payment
	err = service.Reject(payment.ID)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}
}

func TestService_Reject_notFound(t *testing.T) {
	service := &Service{}

	account, err := service.RegisterAccount("+998204567898")
	if err != nil {
		t.Errorf("Can not register an account, error: %v ", err)
		return
	}

	err = service.Deposit(account.ID, 500)
	if err != nil {
		t.Errorf("Can not deposit an account, error: %v", err)
		return
	}

	payment, err := service.Pay(account.ID, 100, "cafe")
	if err != nil {
		t.Errorf("Can not create a payment, error: %v", err)
		return
	}

	// trying to cancel the payment
	changedPaymentID := payment.ID + "2"
	err = service.Reject(changedPaymentID)
	// (if err == nil) ----equivalent---- (if err != ErrPaymentNotFound)
	if err != ErrPaymentNotFound {
		t.Errorf("Error: %v", err)
		return
	}
}
