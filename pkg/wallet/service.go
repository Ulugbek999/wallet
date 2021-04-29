package wallet

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"io"
	"math"
	"github.com/google/uuid"

	"github.com/Ulugbek999/wallet/pkg/types"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var (
	ErrPhoneNumberRegistred = errors.New("phone already registred")
	ErrAmountMustBePositive = errors.New("amount must be greater that zero")
	ErrAccountNotFound      = errors.New("account not found")
	ErrNotEnoughBalance     = errors.New("not enough balance")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrFavoriteNotFound     = errors.New("favorite not found")
	ErrFileNotFound         = errors.New("file not found")
)

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneNumberRegistred
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

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return err
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
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

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	targetPayment, targetAccount, err := s.findPaymentAndAccountByPaymentID(paymentID)
	if err != nil {
		return err
	}

	targetPayment.Status = types.PaymentStatusFail
	targetAccount.Balance += targetPayment.Amount

	return nil
}

func (s *Service) findPaymentAndAccountByPaymentID(paymentID string) (*types.Payment, *types.Account, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, nil, err
	}

	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return nil, nil, err
	}

	return payment, account, nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	targetPayment, targetAccount, err := s.findPaymentAndAccountByPaymentID(paymentID)
	if err != nil {
		return nil, err
	}

	return s.Pay(targetAccount.ID, targetPayment.Amount, targetPayment.Category)
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	targetPayment, targetAccount, err := s.findPaymentAndAccountByPaymentID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: targetAccount.ID,
		Name:      name,
		Amount:    targetPayment.Amount,
		Category:  targetPayment.Category,
	}

	s.favorites = append(s.favorites, favorite)

	return favorite, nil
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	result := ""
	for _, account := range s.accounts {
		result += strconv.Itoa(int(account.ID)) + ";"
		result += string(account.Phone) + ";"
		result += strconv.Itoa(int(account.Balance)) + "|"
	}

	err := actionByFile(path, result)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImportFromFile(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return err
	}

	data := string(byteData)

	splitSlice := strings.Split(data, "|")
	for _, split := range splitSlice {
		if split != "" {
			datas := strings.Split(split, ";")

			id, err := strconv.Atoi(datas[0])
			if err != nil {
				log.Println(err)
				return err
			}

			balance, err := strconv.Atoi(datas[2])
			if err != nil {
				log.Println(err)
				return err
			}

			newAccount := &types.Account{
				ID:      int64(id),
				Phone:   types.Phone(datas[1]),
				Balance: types.Money(balance),
			}

			s.accounts = append(s.accounts, newAccount)
		}
	}

	return nil
}

/*func (s *Service) Export(dir string) error {
	if s.accounts != nil {
		result := ""
		for _, account := range s.accounts {
			result += strconv.Itoa(int(account.ID)) + ";"
			result += string(account.Phone) + ";"
			result += strconv.Itoa(int(account.Balance)) + "\n"
		}

		err := actionByFile(dir+"/accounts.dump", result)
		if err != nil {
			return err
		}
	}

	if s.payments != nil {
		result := ""
		for _, payment := range s.payments {
			result += payment.ID + ";"
			result += strconv.Itoa(int(payment.AccountID)) + ";"
			result += strconv.Itoa(int(payment.Amount)) + ";"
			result += string(payment.Category) + ";"
			result += string(payment.Status) + "\n"
		}

		err := actionByFile(dir+"/payments.dump", result)
		if err != nil {
			return err
		}
	}

	if s.favorites != nil {
		result := ""
		for _, favorite := range s.favorites {
			result += favorite.ID + ";"
			result += strconv.Itoa(int(favorite.AccountID)) + ";"
			result += favorite.Name + ";"
			result += strconv.Itoa(int(favorite.Amount)) + ";"
			result += string(favorite.Category) + "\n"
		}

		err := actionByFile(dir+"/favorites.dump", result)
		if err != nil {
			return err
		}
	}

	return nil
}

*/
/*
func (s *Service) Import(dir string) error {
	err := s.actionByAccounts(dir + "/accounts.dump")
	if err != nil {
		log.Println("err from actionByAccount")
		return err
	}

	err = s.actionByPayments(dir + "/payments.dump")
	if err != nil {
		log.Println("err from actionByPayments")
		return err
	}

	err = s.actionByFavorites(dir + "/favorites.dump")
	if err != nil {
		log.Println("err from actionByFavorites")
		return err
	}

	return nil
} */

func (s *Service) actionByAccounts(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err == nil {
		datas := string(byteData)
		splits := strings.Split(datas, "\n")

		for _, split := range splits {
			if len(split) == 0 {
				break
			}

			data := strings.Split(split, ";")

			id, err := strconv.Atoi(data[0])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			phone := types.Phone(data[1])

			balance, err := strconv.Atoi(data[2])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			account, err := s.FindAccountByID(int64(id))
			if err != nil {
				acc, err := s.RegisterAccount(phone)
				if err != nil {
					log.Println("err from register account")
					return err
				}

				acc.Balance = types.Money(balance)
			} else {
				account.Phone = phone
				account.Balance = types.Money(balance)
			}
		}
	} else {
		log.Println(ErrFileNotFound.Error())
	}

	return nil
}

func (s *Service) actionByPayments(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err == nil {
		datas := string(byteData)
		splits := strings.Split(datas, "\n")

		for _, split := range splits {
			if len(split) == 0 {
				break
			}

			data := strings.Split(split, ";")
			id := data[0]

			accountID, err := strconv.Atoi(data[1])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			amount, err := strconv.Atoi(data[2])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			category := types.PaymentCategory(data[3])

			status := types.PaymentStatus(data[4])

			payment, err := s.FindPaymentByID(id)
			if err != nil {
				newPayment := &types.Payment{
					ID:        id,
					AccountID: int64(accountID),
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(category),
					Status:    types.PaymentStatus(status),
				}

				s.payments = append(s.payments, newPayment)
			} else {
				payment.AccountID = int64(accountID)
				payment.Amount = types.Money(amount)
				payment.Category = category
				payment.Status = status
			}
		}
	} else {
		log.Println(ErrFileNotFound.Error())
	}

	return nil
}

func (s *Service) actionByFavorites(path string) error {
	byteData, err := ioutil.ReadFile(path)
	if err == nil {
		datas := string(byteData)
		splits := strings.Split(datas, "\n")

		for _, split := range splits {
			if len(split) == 0 {
				break
			}

			data := strings.Split(split, ";")
			id := data[0]

			accountID, err := strconv.Atoi(data[1])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			name := data[2]

			amount, err := strconv.Atoi(data[3])
			if err != nil {
				log.Println("can't parse str to int")
				return err
			}

			category := types.PaymentCategory(data[4])

			favorite, err := s.FindFavoriteByID(id)
			if err != nil {
				newFavorite := &types.Favorite{
					ID:        id,
					AccountID: int64(accountID),
					Name:      name,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(category),
				}

				s.favorites = append(s.favorites, newFavorite)
			} else {
				favorite.AccountID = int64(accountID)
				favorite.Name = name
				favorite.Amount = types.Money(amount)
				favorite.Category = category
			}
		}
	} else {
		log.Println(ErrFileNotFound.Error())
	}

	return nil
}

func (s *Service) FindFavoriteByID(id string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == id {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

func actionByFile(path, data string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	_, err = file.WriteString(data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Service) ExportAccountHistory(accountID int64) (payments []types.Payment, err error) {
	_, err = s.FindAccountByID(accountID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			payments = append(payments, *payment)
		}
	}

	if len(payments) == 0 {
		log.Println("empty payment")
		return nil, ErrPaymentNotFound
	}

	return payments, nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) == 0 {
		log.Print(ErrPaymentNotFound)
		return nil
	}

	//log.Printf("payments = %v \n dir = %v \n records = %v", payments, dir, records)

	if len(payments) <= records {
		result := ""
		for _, payment := range payments {
			result += payment.ID + ";"
			result += strconv.Itoa(int(payment.AccountID)) + ";"
			result += strconv.Itoa(int(payment.Amount)) + ";"
			result += string(payment.Category) + ";"
			result += string(payment.Status) + "\n"
		}

		err := actionByFile(dir+"/payments.dump", result)
		if err != nil {
			return err
		}

		return nil
	}

	result := ""
	k := 1
	for i, payment := range payments {
		result += payment.ID + ";"
		result += strconv.Itoa(int(payment.AccountID)) + ";"
		result += strconv.Itoa(int(payment.Amount)) + ";"
		result += string(payment.Category) + ";"
		result += string(payment.Status) + "\n"

		if (i+1)%records == 0 {
			err := actionByFile(dir+"/payments"+strconv.Itoa(k)+".dump", result)
			if err != nil {
				return err
			}
			k++
			result = ""
		}
	}

	if result != "" {
		err := actionByFile(dir+"/payments"+strconv.Itoa(k)+".dump", result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SumPayments(goroutines int) types.Money {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	var summ types.Money = 0
	if goroutines == 0 || goroutines == 1 {
		wg.Add(1)
		go func(payments []*types.Payment) {
			defer wg.Done()
			for _, payment := range payments {
				summ += payment.Amount
			}
		}(s.payments)
	} else {
		from := 0
		count := len(s.payments) / goroutines
		for i := 1; i <= goroutines; i++ {
			wg.Add(1)
			last := len(s.payments) - i*count
			if i == goroutines {
				last = 0
			}
			to := len(s.payments) - last
			go func(payments []*types.Payment) {
				defer wg.Done()
				s := types.Money(0)
				for _, payment := range payments {
					s += payment.Amount
				}
				mu.Lock()
				defer mu.Unlock()
				summ += s
			}(s.payments[from:to])
			from += count
		}
	}

	wg.Wait()

	return summ
}

func (s *Service) FilterPayments(accountID int64, goroutines int) ([]types.Payment, error) {
	filteredPayments := []types.Payment{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	if goroutines == 0 || goroutines == 1 {
		wg.Add(1)
		go func(payments []*types.Payment) {
			defer wg.Done()
			for _, payment := range payments {
				if payment.AccountID == accountID {
					filteredPayments = append(filteredPayments, types.Payment{
						ID:        payment.ID,
						AccountID: payment.AccountID,
						Amount:    payment.Amount,
						Category:  payment.Category,
						Status:    payment.Status,
					})
				}
			}
		}(s.payments)
	} else {
		from := 0
		count := len(s.payments) / goroutines
		for i := 1; i <= goroutines; i++ {
			wg.Add(1)
			last := len(s.payments) - i*count
			if i == goroutines {
				last = 0
			}
			to := len(s.payments) - last
			go func(payments []*types.Payment) {
				defer wg.Done()
				separetePayments := []types.Payment{}
				for _, payment := range payments {
					if payment.AccountID == accountID {
						separetePayments = append(separetePayments, types.Payment{
							ID:        payment.ID,
							AccountID: payment.AccountID,
							Amount:    payment.Amount,
							Category:  payment.Category,
							Status:    payment.Status,
						})
					}
				}
				mu.Lock()
				defer mu.Unlock()
				filteredPayments = append(filteredPayments, separetePayments...)
			}(s.payments[from:to])
			from += count
		}
	}

	wg.Wait()

	if len(filteredPayments) == 0 {
		return nil, ErrAccountNotFound
	}

	return filteredPayments, nil
}



// homework 17

func (s *Service) Export(dir string) error {

	// dir, err := filepath.Abs(dir)

	// if err != nil {
	// 	log.Print(err)

	// }
	var err = errors.New("Error")
	err = nil
	lenAcou := len(s.accounts)

	if lenAcou != 0 {

		dirAccount := dir + "/accounts.dump"
		//	log.Print(dirAccount)

		fileAccounts, err := os.Create(dirAccount)
		if err != nil {
			log.Print(err)

		}

		defer func() {
			if cerr := fileAccounts.Close(); err != nil {
				log.Print(cerr)
			}
		}()

		for index, account := range s.accounts {
			//	account, err = s.FindAccountByID(int64(ind))
			// fmt.Println(newP2)
			// fmt.Println(ee3)
			if index != 0 {
				_, err = fileAccounts.Write([]byte("\n"))
				if err != nil {
					log.Print(err)

				}

			}
			_, err = fileAccounts.Write([]byte(strconv.FormatInt((account.ID), 10)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileAccounts.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}
			_, err = fileAccounts.Write([]byte(string(account.Phone)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileAccounts.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = fileAccounts.Write([]byte(strconv.FormatInt(int64(account.Balance), 10)))
			if err != nil {
				log.Print(err)

			}

		}
	}

	lenPay := len(s.payments)
	if lenPay != 0 {

		dirPayment := dir + "/payments.dump"
		filePayments, err := os.Create(dirPayment)
		if err != nil {
			log.Print(err)

		}

		defer func() {
			if cerr := filePayments.Close(); err != nil {
				log.Print(cerr)
			}
		}()

		for index, payment := range s.payments {
			//	account, err = s.FindAccountByID(int64(ind))
			// fmt.Println(newP2)

			if index != 0 {
				_, err = filePayments.Write([]byte("\n"))
				if err != nil {
					log.Print(err)

				}

			}

			_, err = filePayments.Write([]byte(string(payment.ID)))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(strconv.FormatInt(int64(payment.AccountID), 10)))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(strconv.FormatInt(int64(payment.Amount), 10)))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(string(payment.Category)))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = filePayments.Write([]byte(string(payment.Status)))
			if err != nil {
				log.Print(err)

			}

		}
	}

	lenFav := len(s.favorites)

	if lenFav != 0 {
		dirFavorite := dir + "/favorites.dump"
		fileFavorites, err := os.Create(dirFavorite)
		if err != nil {
			log.Print(err)

		}

		defer func() {
			if cerr := fileFavorites.Close(); err != nil {
				log.Print(cerr)
			}
		}()

		for index, favorite := range s.favorites {
			//	account, err = s.FindAccountByID(int64(ind))
			// fmt.Println(newP2)

			if index != 0 {
				_, err = fileFavorites.Write([]byte("\n"))
				if err != nil {
					log.Print(err)

				}

			}

			_, err = fileFavorites.Write([]byte(string(favorite.ID)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(strconv.FormatInt(int64(favorite.AccountID), 10)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(string(favorite.Name)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(strconv.FormatInt(int64(favorite.Amount), 10)))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(";"))
			if err != nil {
				log.Print(err)

			}

			_, err = fileFavorites.Write([]byte(string(favorite.Category)))
			if err != nil {
				log.Print(err)

			}

		}
	}

	return err

}

//Import for
func (s *Service) Import(dir string) error {

	dirAccount := dir + "/accounts.dump"
	fileAccount, err := os.Open(dirAccount)
	//	log.Print(dirAccount)
	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			err := fileAccount.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		//		log.Printf("%#v", fileAccount)

		content := make([]byte, 0)
		buf := make([]byte, 4)
		for {
			read, err := fileAccount.Read(buf)
			if err == io.EOF {
				break
			}
			content = append(content, buf[:read]...)
		}

		data := string(content)
		newData := strings.Split(data, "\n")
		//log.Print(data)
		//log.Print(newData)

		for _, stroka := range newData {
			//log.Print(stroka)
			account := &types.Account{}
			newData2 := strings.Split(stroka, ";")
			for ind, stroka2 := range newData2 {
				// if stroka2 == "" {
				// 	return ErrPhoneRegistered
				// }
				//log.Print(stroka2)
				if ind == 0 {
					id, _ := strconv.ParseInt(stroka2, 10, 64)
					account.ID = id
				}
				if ind == 1 {
					account.Phone = types.Phone(stroka2)
				}
				if ind == 2 {
					balance, _ := strconv.ParseInt(stroka2, 10, 64)
					account.Balance = types.Money(balance)

				}

				//	log.Print(ind1)

			}
			errExist := 1
			for _, accountCheck := range s.accounts {

				if accountCheck.ID == account.ID {
					accountCheck.Phone = account.Phone
					accountCheck.Balance = account.Balance
					errExist = 0
				}

			}
			if errExist == 1 {
				s.accounts = append(s.accounts, account)
			}
		}

	}

	dirPayment := dir + "/payments.dump"
	filePayments, err := os.Open(dirPayment)
	if err != nil {
		log.Print(err)
		return ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			err := filePayments.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		//		log.Printf("%#v", filePayments)

		contentPayment := make([]byte, 0)
		bufPayment := make([]byte, 4)
		for {
			read, err := filePayments.Read(bufPayment)
			if err == io.EOF {
				break
			}
			contentPayment = append(contentPayment, bufPayment[:read]...)
		}

		dataPayment := string(contentPayment)
		newDataPayment := strings.Split(dataPayment, "\n")
		//log.Print(data)
		//log.Print(newData)

		for _, stroka := range newDataPayment {
			//log.Print(stroka)
			payment := &types.Payment{}
			newData2 := strings.Split(stroka, ";")
			for ind, stroka2 := range newData2 {
				// if stroka2 == "" {
				// 	return ErrPhoneRegistered
				// }
				//log.Print(stroka2)
				if ind == 0 {
					//id, _ := stroka2
					payment.ID = stroka2
				}
				if ind == 1 {
					accountID, _ := strconv.ParseInt(stroka2, 10, 64)
					payment.AccountID = int64(accountID)
				}

				if ind == 2 {
					balance, _ := strconv.ParseInt(stroka2, 10, 64)
					payment.Amount = types.Money(balance)
				}

				if ind == 3 {
					payment.Category = types.PaymentCategory(stroka2)
				}

				if ind == 4 {
					payment.Status = types.PaymentStatus(stroka2)
				}

				//		log.Print(ind1)

			}
			errExist := 1
			for _, paymentCheck := range s.payments {

				if paymentCheck.ID == payment.ID {
					paymentCheck.AccountID = payment.AccountID
					paymentCheck.Amount = payment.Amount
					paymentCheck.Category = payment.Category
					paymentCheck.Status = payment.Status
					errExist = 0
				}

			}
			if errExist == 1 {
				s.payments = append(s.payments, payment)
			}
		}

	}

	dirFavorite := dir + "/favorites.dump"
	fileFavorites, err := os.Open(dirFavorite)
	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}
	if err != ErrFileNotFound {
		defer func() {
			err := fileFavorites.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		//	log.Printf("%#v", fileFavorites)

		contentFavorite := make([]byte, 0)
		bufFavorite := make([]byte, 4)
		for {
			read, err := fileFavorites.Read(bufFavorite)
			if err == io.EOF {
				break
			}
			contentFavorite = append(contentFavorite, bufFavorite[:read]...)
		}

		dataFavorite := string(contentFavorite)
		newDataFavorite := strings.Split(dataFavorite, "\n")
		//log.Print(data)
		//log.Print(newData)

		for _, stroka := range newDataFavorite {
			//log.Print(stroka)
			favorite := &types.Favorite{}
			newData2 := strings.Split(stroka, ";")
			for ind, stroka2 := range newData2 {
				// if stroka2 == "" {
				// 	return ErrPhoneRegistered
				// }
				//log.Print(stroka2)
				if ind == 0 {
					//id, _ := stroka2
					favorite.ID = stroka2
				}
				if ind == 1 {
					accountID, _ := strconv.ParseInt(stroka2, 10, 64)
					favorite.AccountID = int64(accountID)
				}

				if ind == 2 {
					favorite.Name = stroka2
				}
				if ind == 3 {
					balance, _ := strconv.ParseInt(stroka2, 10, 64)
					favorite.Amount = types.Money(balance)
				}

				if ind == 4 {
					favorite.Category = types.PaymentCategory(stroka2)
				}

				//	log.Print(ind1)

			}
			errExist := 1
			for _, favoriteCheck := range s.favorites {

				if favoriteCheck.ID == favorite.ID {
					favoriteCheck.AccountID = favorite.AccountID
					favoriteCheck.Name = favorite.Name
					favoriteCheck.Amount = favorite.Amount
					favoriteCheck.Category = favorite.Category
					errExist = 0
				}

			}
			if errExist == 1 {
				s.favorites = append(s.favorites, favorite)
			}
		}

	}
	return nil

}


//FilterPaymentsByFn for
func (s *Service) FilterPaymentsByFn(filter func(payment types.Payment) bool, goroutines int) ([]types.Payment, error) {
	var foundPayments []types.Payment
	
	var allfoundPayments []types.Payment
	for _, payment := range s.payments {

		if filter(*payment) == true {
			foundPayments = append(foundPayments, *payment)
			
		}
	}
	if foundPayments == nil {
		return nil, ErrAccountNotFound
	}

	if goroutines <= 1 {
		return foundPayments, nil
	}

	wg := sync.WaitGroup{}

	wg.Add(goroutines) 

	mu := sync.Mutex{}

	lenPay := len(foundPayments)
	numberOfPaymentPerRoutine := 0

	numberOfPaymentPerRoutine = int(math.Ceil(float64((lenPay + 1) / goroutines)))

	

	index := 0
	newNumberOfPaymentPerRoutine := numberOfPaymentPerRoutine

	go func() {
		for i := 0; i < goroutines; i++ {
			lenPay := len(foundPayments)

			

			defer wg.Done() 
			var newPayments []types.Payment

			for i := 0; index < numberOfPaymentPerRoutine; i++ {
				
				newPayments = append(newPayments, foundPayments[index])
				
				index++
			}

			mu.Lock()
			numberOfPaymentPerRoutine += newNumberOfPaymentPerRoutine

			allfoundPayments = append(allfoundPayments, newPayments...)
			mu.Unlock()
			if (i == goroutines-1) && (len(allfoundPayments) != lenPay) {

				foundLen := len(allfoundPayments)
				for j := foundLen; j < lenPay; j++ {
					allfoundPayments = append(allfoundPayments, foundPayments[j])
				}
			}

		}
	}()


	wg.Wait()
	return allfoundPayments, nil
}



type Progress struct {
	Part int
	Result types.Money
  }







//SumPaymentsWithProgress ..
func (s *Service) SumPaymentsWithProgress() <-chan Progress {
	if len(s.payments) == 0 {
		ch := make(chan Progress)
		close(ch)
		return ch
	}

	const size = 100_000
	var parts int = len(s.payments) / size
	channels := make([]<-chan Progress, parts)

	for i := 0; i <= parts; i++ {
		//TODO здес будем запускат горутины
		left := i * size
		right := (i + 1) * size
		if right > len(s.payments) {
			right = len(s.payments)
		}
		ch := make(chan Progress) //здесь вы создаете канал для каждого горутину
		go func(ch chan<- Progress, data []*types.Payment) {
			defer close(ch)
			total := types.Money(0)
			for _, v := range data {
				total += v.Amount
			}
			ch <- Progress{
				Part:   len(data),
				Result: total,
			}
		}(ch, s.payments[left:right])
		channels[i] = ch
	}
	return merge(channels)
}

func merge(channels []<-chan Progress) <-chan Progress {
	wg := sync.WaitGroup{}
	wg.Add(len(channels))
	merged := make(chan Progress)
	for _, ch := range channels {
		go func(ch <-chan Progress) {
			defer wg.Done()
			for val := range ch {
				merged <- val
			}
		}(ch)
	}
	go func() {
		defer close(merged)
		wg.Wait()
	}()
	return merged
}