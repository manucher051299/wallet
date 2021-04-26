package wallet

import (
	"bufio"
	"errors"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/manucher051299/wallet/pkg/types"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) SumPayments(goroutines int) types.Money {
	if goroutines <= 1 {
		sum := 0
		for _, payment := range s.payments {
			sum += int(payment.Amount)
		}
		return types.Money(sum)
	}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	sum := int64(0)
	wg.Add(goroutines)
	sliceLen := int(math.Ceil(float64(len(s.payments)) / float64(goroutines)))
	for i := 0; i < len(s.payments); i += sliceLen {
		if i+sliceLen > len(s.payments) {
			sliceLen = len(s.payments) - i
		}
		go func(j int, len int) {
			defer wg.Done()

			mu.Lock()
			for ; j < len; j++ {
				sum += int64(s.payments[j].Amount)
			}
			mu.Unlock()
		}(i, sliceLen)
	}
	return types.Money(sum)
}
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	payments := make([]types.Payment, 0)
	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			payments = append(payments, *payment)
		}
	}
	return payments, nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	var file *os.File
	var err error
	if len(payments) == 0 {
		return nil
	}
	if len(payments) <= records {
		file, err = os.Create(dir + "/payments.dump")
		if err != nil {
			return err
		}
	} else {
		file, err = os.Create(dir + "/payments1.dump")
		if err != nil {
			return err
		}
	}
	x := 1
	i := 1
	for _, payment := range payments {
		log.Println(strconv.Itoa(i) + " " + strconv.Itoa(x) + " " + strconv.Itoa(records))
		if i%records == 1 && i != 1 {
			x++
			file, err = os.Create(dir + "/payments" + strconv.Itoa(x) + ".dump")
			if err != nil {
				return err
			}
		}
		_, err := file.Write([]byte(payment.ID + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(payment.AccountID, 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(int64(payment.Amount), 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(payment.Category + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(payment.Status + "\n"))
		if err != nil {
			log.Print(err)
			return err
		}
		i++
	}
	return nil
}
func (s *Service) Export(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if s.accounts == nil {
		return nil
	}
	file, err := os.Create(dir + "/accounts.dump")
	if err != nil {
		return err
	}
	for _, account := range s.accounts {
		_, err := file.Write([]byte(strconv.FormatInt(account.ID, 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(account.Phone + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(int64(account.Balance), 10) + "\n"))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	if s.payments == nil {
		return nil
	}
	file, err = os.Create(dir + "/payments.dump")
	for _, payment := range s.payments {
		_, err := file.Write([]byte(payment.ID + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(payment.AccountID, 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(int64(payment.Amount), 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(payment.Category + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(payment.Status + "\n"))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	if s.favorites == nil {
		return nil
	}
	file, err = os.Create(dir + "/favorites.dump")
	for _, favorite := range s.favorites {
		_, err := file.Write([]byte(favorite.ID + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(favorite.AccountID, 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(strconv.FormatInt(int64(favorite.Amount), 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(favorite.Name + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		_, err = file.Write([]byte(favorite.Category + "\n"))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}
func (s *Service) Import(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	srcAcc, err := os.Open(dir + "/accounts.dump")
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := srcAcc.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()
	reader := bufio.NewReader(srcAcc)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		account := strings.Split(strings.Split(line, "\n")[0], ";")
		id, err := strconv.ParseInt(account[0], 10, 64)
		if err != nil {
			return err
		}
		balance, err := strconv.ParseInt(account[2], 10, 64)
		if err != nil {
			return err
		}
		s.accounts = append(s.accounts, &types.Account{ID: id, Balance: types.Money(balance), Phone: types.Phone(account[1])})
	}
	srcPay, err := os.Open(dir + "/payments.dump")
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := srcPay.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()
	reader = bufio.NewReader(srcPay)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		log.Printf(line)
		if err != nil {
			log.Print(err)
			return err
		}
		payment := strings.Split(strings.Split(line, "\n")[0], ";")
		AccId, err := strconv.ParseInt(payment[1], 10, 64)
		if err != nil {
			return err
		}
		amount, err := strconv.ParseInt(payment[2], 10, 64)
		if err != nil {
			return err
		}
		s.payments = append(s.payments, &types.Payment{ID: payment[0], AccountID: AccId, Amount: types.Money(amount), Category: types.PaymentCategory(payment[3]), Status: types.PaymentStatus(payment[4])})
	}

	srcFav, err := os.Open(dir + "/favorites.dump")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := srcFav.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()
	reader = bufio.NewReader(srcFav)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		favorite := strings.Split(strings.Split(line, "\n")[0], ";")
		AccId, err := strconv.ParseInt(favorite[1], 10, 64)
		if err != nil {
			return err
		}
		amount, err := strconv.ParseInt(favorite[2], 10, 64)
		if err != nil {
			return err
		}
		s.favorites = append(s.favorites, &types.Favorite{ID: favorite[0], AccountID: AccId, Amount: types.Money(amount), Name: favorite[3], Category: types.PaymentCategory(favorite[4])})
	}
	return nil
}
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	for _, acc := range s.accounts {
		_, err = file.Write([]byte(strconv.FormatInt(acc.ID, 10) + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		file.Write([]byte(acc.Phone + ";"))
		if err != nil {
			log.Print(err)
			return err
		}
		file.Write([]byte(strconv.FormatInt(int64(acc.Balance), 10) + "|"))
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err = file.Close()
	if err != nil {
		log.Print(err)
	}
	return err
}
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	buff := make([]byte, 4086)
	content := make([]byte, 0)
	for {
		read, err := file.Read(buff)
		if err == io.EOF {
			content = append(content, buff[:read]...)
			break
		}
		if err != nil {
			return err
		}
		content = append(content, buff[:read]...)
	}
	data := strings.Split(string(content), "|")
	for _, acc := range data {
		if acc != "" {
			account := strings.Split(acc, ";")
			id, err := strconv.ParseInt(account[0], 10, 64)
			if err != nil {
				return err
			}
			balance, err := strconv.ParseInt(account[1], 10, 64)
			if err != nil {
				return err
			}
			phone := account[1]
			s.accounts = append(s.accounts, &types.Account{ID: id, Balance: types.Money(balance), Phone: types.Phone(phone)})
		}
	}
	err = file.Close()
	if err != nil {
		log.Print(err)
	}
	return err
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
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// зачисление средств пока не рассматриваем как платёж
	account.Balance += amount
	return nil
}

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

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	return s.Pay(payment.AccountID, payment.Amount, payment.Category)
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Name:      name,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}
	return s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
}
