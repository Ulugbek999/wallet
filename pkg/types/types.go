package types

type Money int64

type PaymentCategory string

type PaymentStatus string

const (
	PaymentStatusOk PaymentStatus = "OK"
	PaymentStatusFail PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

type Payment struct {
	ID string
	accountID int64
	Amount Money
	Catergory PaymentCategory
	Status PaymentStatus
}

type Phone string

type Account struct {
	ID int
	Phone Phone
	Balance Money
}