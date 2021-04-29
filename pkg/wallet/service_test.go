package wallet


import (
	"fmt"
	"github.com/Ulugbek999/wallet/pkg/types"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

type testService struct {
	*Service
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992938151007",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{{
		amount:   1000_00,
		category: "auto",
	}},
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)

	if err != nil {
		return nil, fmt.Errorf("cant register account, error = %v", err)
	}

	err = s.Deposit(account.ID, balance)

	if err != nil {
		return nil, fmt.Errorf("cant deposit account, error = %v", err)
	}

	return account, nil
}

func (s *testService) addAcoount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("cant register account %v = ", err)
	}

	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("cant deposit account %v = ", err)
	}

	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("cant make payment %v = ", err)
		}
	}

	return account, payments, nil
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]

	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentId(), error %v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentId(), wrong payment returned %v = ", got)
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()

	_, _, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentId(), must return error, returned nil %v = ", err)
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentId(), must return ErrPaymentNotFound returned %v = ", err)
	}
}

func TestService_FindbyAccountById_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9929351007")
	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("не удалось найти аккаунт, получили: %v", account)
	}

}

func TestService_FindByAccountByID_notFound(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+992938151007")
	account, err := svc.FindAccountByID(2)
	if err == nil {
		t.Errorf("аккаунт найден, аккаунт: %v", account)
	}

}

func TestService_Reject_success(t *testing.T) {
	// создаем сервис
	s := newTestService()

	_, payments, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	err = s.Reject(payment.ID)

	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): cant find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didnt changed, payment = %v", savedPayment)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): cant find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didnt changed, balance = %v", savedAccount)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {

	s := newTestService()

	_, payments, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	newPayment, err := s.Repeat(payment.ID)

	if err != nil {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	if newPayment.AccountID != payment.AccountID {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	if newPayment.Amount != payment.Amount {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	if newPayment.Category != payment.Category {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

	if newPayment.Status != payment.Status {
		t.Errorf("Repeat(): error = %v", err)
		return
	}

}

func TestService_FindFavoriteByID_succes(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	favorite, err := s.FavoritePayment(payment.ID, "ogastus")
	if err != nil {
		t.Errorf("favorite cant find, error = %v", err)
	}

	paymentFavorite, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v), error = %v", paymentFavorite, err)
	}
}

func TestService_Export_success(t *testing.T) {
	s := newTestService()

	_, payments, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, payments, err = s.addAcoount(testAccount{
		phone:   "+992938151003",
		balance: 10_000_00,
		payments: []struct {
			amount   types.Money
			category types.PaymentCategory
		}{{
			amount:   1000_00,
			category: "auto",
		}},
	})
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	favorite, err := s.FavoritePayment(payment.ID, "ogastus")
	if err != nil {
		t.Errorf("cant find favorite, error = %v", err)
	}

	paymentFavorite, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite() can't for an favorite(%v), error = %v", paymentFavorite, err)
	}

	err = s.Export("data")
	if err != nil {
		t.Errorf("Export() Error can't export error = %v", err)
	}
}

func TestService_Import_success(t *testing.T) {
	s := newTestService()

	err := s.Import("data")

	if err != nil {
		t.Errorf("Import() Error can't import error = %v", err)
	}
}

func TestService_HistoryToFiles_success(t *testing.T) {
	s := newTestService()

	_, _, err := s.addAcoount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	_, _, err = s.addAcoount(testAccount{
		phone:   "+992938151008",
		balance: 10_000_00,
		payments: []struct {
			amount   types.Money
			category types.PaymentCategory
		}{{
			amount:   1000_00,
			category: "auto",
		}, {
			amount:   1020_00,
			category: "auto",
		},
			{
				amount:   1030_00,
				category: "auto",
			}},
	})

	if err != nil {
		t.Error(err)
		return
	}

	payments, err := s.ExportAccountHistory(2)
	if err != nil {
		t.Error(err)
		return
	}

	err = s.HistoryToFiles(payments, "data", 2)
	if err != nil {
		t.Errorf("HistoryToFiles() Error can't export to file, error = %v", err)
		return
	}
}

func BenchmarkSumPayments(b *testing.B) {
	s := newTestService()

	_, _, err := s.addAcoount(testAccount{
		phone:   "+992938151002",
		balance: 10_000,
		payments: []struct {
			amount   types.Money
			category types.PaymentCategory
		}{
			{
				amount:   100,
				category: "auto",
			},
			{
				amount:   200,
				category: "auto",
			},
			{
				amount:   300,
				category: "auto",
			},
			{
				amount:   400,
				category: "auto",
			},
			{
				amount:   500,
				category: "auto",
			},
			{
				amount:   600,
				category: "auto",
			},
			{
				amount:   100,
				category: "auto",
			},
			{
				amount:   100,
				category: "auto",
			},
			{
				amount:   100,
				category: "auto",
			},
			{
				amount: 100,
				category: "auto",
			},
		},
	})

	if err != nil {
		b.Error(err)
		return
	}

	want := types.Money(2500)

	for i := 0; i < b.N; i++ {
		// на каждый горутине по 2 слайса
		result := s.SumPayments(100)
		if result != want {
			b.Fatalf("invalid result, got = %v want = %v", result, want)
		}
	}
}


func TestService_SumPaymentsWithProgress(t *testing.T) {
	s := newTestService()
	for i := 0; i < 200_000; i++ {
		payment := &types.Payment{
			ID:     uuid.New().String(),
			Amount: types.Money(100),
		}
		s.payments = append(s.payments, payment)
	}

	s.SumPaymentsWithProgress()
	
}