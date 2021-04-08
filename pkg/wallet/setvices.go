package wallet

import (
	"github.com/Ulugbek999/wallet/pkg/types"
)




type Service struct {
	nextAccountID int64
	accounts []types.Account
	payments []types.Payment

}


func RegisterAccount(service *Service, phone types.Phone)  {
	
	for _, account := range service.accounts {
		if account.Phone == phone{
			return
		}
	}

	service.nextAccountID++
	service.accounts = append(service.accounts, types.Account{
		ID: int(service.nextAccountID),
		Phone: phone,
		Balance: 0,
	})


}

func (service *Service)RegisterAccount( phone types.Phone)  {
	
	for _, account := range service.accounts {
		if account.Phone == phone{
			return
		}
	}

	service.nextAccountID++
	service.accounts = append(service.accounts, types.Account{
		ID: int(service.nextAccountID),
		Phone: phone,
		Balance: 0,
	})


}



















/*func FindAccountByID(accountID int64) (*types.Account, error) {







}*/


/*
//CategoriesAvg расчитовает среднюю сумму платежа
func CategoriesAvg(payments []types.Payment) map[types.Category]types.Money {
	count := map[types.Category]types.Money{}
	result := map[types.Category]types.Money{}

	for _, payment := range payments {
		result[payment.Category] += payment.Amount
		count[payment.Category]++
	}

	for key := range result {
		result[key] /= count[key]
	}

	return result
}




//PeriodsDynamic расчитовает сумму платежа по категории
func PeriodsDynamic(first map[types.Category]types.Money, second map[types.Category]types.Money) map[types.Category]types.Money {

	amount := map[types.Category]types.Money{}

	for sum := range second {
		amount[sum] += second[sum]
	}

	for sum := range first {
		amount[sum] -= first[sum]
	}

	return amount
}

*/