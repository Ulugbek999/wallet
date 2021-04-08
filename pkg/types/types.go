package types

//Money - represents a monetary amount
//in minimum units (cents, kopecks, diramas, etc.).
type Money int64

//PaymentCategory - represents the category in which
//the payment was made(auto, pharmacies, restaurants etc.).
type PaymentCategory string

//PaymentStatus - represents the status of the payments
type PaymentStatus string

//Predefined payment statuses
const (
	PaymentStatusOK         PaymentStatus = "OK"
	PaymentStatusFail       PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

//Payment - represents information about the payment source.
type Payment struct {
	ID        string
	AccountID int64
	Amount    Money
	Category  PaymentCategory
	Status    PaymentStatus
}

//Phone - phone number
type Phone string

//
type Account struct {
	ID      int64
	Phone   Phone
	Balance Money
}