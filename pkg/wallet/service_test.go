package wallet

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/manucher051299/wallet/pkg/types"
)

func BenchmarkSumPayments_single(b *testing.B) {
	s := newTestService()
	s.addAccount(defaultTestAccount)
	want := types.Money(1000_00)
	for i := 0; i < b.N; i++ {
		sum := s.SumPayments(1)
		if sum != want {
			b.Fatalf("invalid result, got %v wanted %v", sum, want)
		}
	}
}
func BenchmarkSumPayments_multiple(b *testing.B) {
	s := newTestService()
	s.addAccount(defaultTestAccount2)
	want := types.Money(2000_00)
	for i := 0; i < b.N; i++ {
		sum := s.SumPayments(2)
		if sum != want {
			b.Fatalf("invalid result, got %v wanted %v", sum, want)
		}
	}
}
func TestService_ExportImport(t *testing.T) {
	s := newTestService()
	s.addAccount(defaultTestAccount)
	err := s.Export("../../data")
	if err != nil {
		t.Error(err)
	}
	s2 := Service{}
	err = s2.Import("../../data")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(s.accounts); i++ {
		if *s.accounts[i] != *s2.accounts[i] {
			log.Printf("exported: ")
			log.Println(s.accounts[i])
			log.Printf("imported: ")
			log.Println(s2.accounts[i])
			t.Errorf("accounts not same")
			return
		}
	}
	for i := 0; i < len(s.payments); i++ {
		if *s.payments[i] != *s2.payments[i] {
			t.Errorf("payments not same")
			return
		}
	}
	for i := 0; i < len(s.favorites); i++ {
		if *s.favorites[i] != *s2.favorites[i] {
			t.Errorf("favorites not same")
			return
		}
	}
}
func TestService_ExportFromFileImportFromFile(t *testing.T) {
	s := newTestService()
	s.addAccount(defaultTestAccount)
	err := s.ExportToFile("../../data/test.txt")
	if err != nil {
		t.Error(err)
	}
	s2 := Service{}
	err = s2.ImportFromFile("../../data/test.txt")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(s.accounts); i++ {
		if *s.accounts[i] != *s2.accounts[i] {
			log.Printf("exported: ")
			log.Println(s.accounts[i])
			log.Printf("imported: ")
			log.Println(s2.accounts[i])
			t.Errorf("accounts not same")
			return
		}
	}
}
func TestService_Register_success(t *testing.T) {
	s := &Service{}
	_, err := s.RegisterAccount("92087397")
	if err != nil {
		t.Error(err)
		return
	}
}
func TestService_FindAccountByID_success(t *testing.T) {
	s := newTestService()
	account, err := s.RegisterAccount("92087397")
	if err != nil {
		t.Error(err)
		return
	}
	account, err = s.FindAccountByID(account.ID)
}
func TestService_Deposit_success(t *testing.T) {
	s := &Service{}
	account, err := s.RegisterAccount("92087397")
	if err != nil {
		t.Error(err)
		return
	}
	err = s.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Error(err)
		return
	}
}
func TestService_Reject_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(), cant reject payment, error = %v", err)
		return
	}
}
func TestService_Reject_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.Reject("asdc")
	if err == nil {
		t.Errorf("Reject(), shoud show error -> error = %v", err)
		return
	}
}
func TestService_Pay_success(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.Pay(account.ID, 10, "auto")
	if err != nil {
		t.Errorf("Reject(), cant reject payment, error = %v", err)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	_, err = s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Reject(), cant reject payment, error = %v", err)
		return
	}
}
func TestService_FindPaymentById_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentById(), cant find payment, error = %v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentById(), wrong payment, error = %v", err)
		return
	}
}

func TestService_FindPaymentById_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentById(), must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentById(), must return ErrPaymentNotFound, error = %v", err)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error((err))
		return
	}
	payment, err := s.Pay(account.ID, 10, "smth")
	if err != nil {
		t.Error(err)
		return
	}
	favoritePayment, err := s.FavoritePayment(payment.ID, "good")
	if err != nil {
		t.Error(err)
		return
	}
	payment, err = s.PayFromFavorite(favoritePayment.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	s := newTestService()
	account, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		return
	}
	payment, err := s.Pay(account.ID, account.Balance+100, "smth")
	if err != nil {
		return
	}
	favoritePayment, err := s.FavoritePayment(payment.ID, "good")
	if err != nil {
		return
	}
	payment, err = s.PayFromFavorite(favoritePayment.ID)
	if err != nil {
		return
	}
	t.Errorf("ERROR, should havefound any error")
	return
}

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
		return nil, fmt.Errorf("can't deposit to account, error = %v", err)
	}
	return account, nil
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, err: %v", err)
	}
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, err: %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, err: %v", err)
		}
	}
	return account, payments, nil
}

var defaultTestAccount = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1000_00, category: "auto"},
	},
}

var defaultTestAccount2 = testAccount{
	phone:   "+992000000001",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1000_00, category: "auto"},
		{amount: 1000_00, category: "smth"},
	},
}
