package psql

import (
	"Etog/internal/domain/entity"
	"testing"
	"time"
)

func setup() string {
	storage_path := "postgresql://postgres:0@localhost:5433/etog?sslmode=disable"
	return storage_path
}

func Test_New(t *testing.T) {
	storage_path := setup()
	storage := New(storage_path, nil)
	if storage == nil {
		t.Errorf("storage is nil")
		return
	}
	t.Log("func [New] is OK")
}

func TestStorage_CRUDAccount(t *testing.T) {
	storage_path := setup()
	storage := New(storage_path, nil)
	if storage == nil {
		t.Errorf("storage is nil")
		return
	}
	account := entity.AccountDb{
		Id:          1,
		Mail:        "some@mail.ru",
		Login:       "izame",
		Password:    "123456",
		Avatar:      "none",
		Official:    false,
		Description: "new test account",
		Rating:      100,
		Deleted:     false,
		Active:      false,
		Followers:   20,
	}
	upd_account := entity.AccountDb{
		Id:          1,
		Mail:        "some@mail.ru",
		Login:       "izame",
		Password:    "123456",
		Avatar:      "none",
		Official:    false,
		Description: "new test account",
		Rating:      100,
		Deleted:     false,
		Active:      false,
		Followers:   25,
	}
	err := storage.CreateAccount(account)
	if err != nil {
		t.Error("ошибка создания аккаунта 1")
		return
	}
	new_account, err := storage.GetAccount(1)
	if err != nil {
		t.Error(err)
		return
	}
	if new_account.Id != account.Id || new_account.Mail != account.Mail || new_account.Login != account.Login {
		t.Errorf("Account id does not match")
		return
	}
	new_account, err = storage.GetAccountByLogin(account.Login)
	if err != nil {
		t.Error(err)
		return
	}
	if new_account.Id != account.Id || new_account.Mail != account.Mail {
		t.Errorf("Account id does not match")
		return
	}
	new_account, err = storage.GetAccountByEmail(account.Mail)
	if err != nil {
		t.Error(err)
		return
	}
	if new_account.Id != account.Id || new_account.Mail != account.Mail {
		t.Errorf("Account id does not match")
		return
	}
	err = storage.PutAccount(upd_account)
	if err != nil {
		t.Error(err)
		return
	}
	new_account, err = storage.GetAccount(1)
	if err != nil {
		t.Error(err)
		return
	}
	if new_account.Followers != upd_account.Followers {
		t.Errorf("Followers does not match")
		return
	}
	storage.db.Where("login = ?", account.Login).Delete(&entity.AccountDb{})
}

func Test_CRUDRefreshToken(t *testing.T) {
	storage_path := setup()
	storage := New(storage_path, nil)
	if storage == nil {
		t.Errorf("storage is nil")
		return
	}
	str_id := "wedwefshs"
	id := []byte(str_id)
	token := entity.RefreshToken{
		TokenId:  id,
		UserId:   21,
		ExpireAt: time.Now().Add(1 * time.Minute),
	}
	err := storage.CreateRefreshToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	db_token, err := storage.GetRefreshToken(id)
	if err != nil {
		t.Error(err)
		return
	}
	dbTime := db_token.ExpireAt.UTC().Truncate(time.Microsecond)
	expectedTime := token.ExpireAt.UTC().Truncate(time.Microsecond)

	if db_token.UserId != token.UserId || dbTime.Equal(expectedTime) {
		t.Errorf("Token id does not match")
		return
	}
	err = storage.DeleteRefreshToken(token.UserId)
	if err != nil {
		t.Error(err)
		return
	}
	db_token, err = storage.GetRefreshToken(id)
	if err == nil {
		t.Error("must be deleted")
		return
	}
}

func Test_CreateOfficialRequest(t *testing.T) {
	storage_path := setup()
	storage := New(storage_path, nil)
	if storage == nil {
		t.Errorf("storage is nil")
		return
	}

	account := entity.AccountDb{
		Id:    999,
		Mail:  "official@test.ru",
		Login: "official_test",
	}
	storage.CreateAccount(account)
	defer storage.db.Where("id = ?", account.Id).Delete(&entity.AccountDb{})

	err := storage.CreateOfficialRequest(account.Id)
	if err != nil {
		t.Errorf("ошибка создания official request: %v", err)
		return
	}

	var req entity.OfficialRequest
	result := storage.db.Where("user_id = ?", account.Id).First(&req)
	if result.Error != nil {
		t.Errorf("official request не найден в БД: %v", result.Error)
		return
	}
	if req.UserID != account.Id {
		t.Errorf("UserID не совпадает: got %d, want %d", req.UserID, account.Id)
		return
	}
	if req.Comment != nil {
		t.Errorf("Comment должен быть nil, got %v", req.Comment)
		return
	}

	// Cleanup
	storage.db.Where("user_id = ?", account.Id).Delete(&entity.OfficialRequest{})
	t.Log("func [CreateOfficialRequest] is OK")
}

func Test_SubscribeUnsubscribe(t *testing.T) {
	storage_path := setup()
	storage := New(storage_path, nil)
	if storage == nil {
		t.Errorf("storage is nil")
		return
	}

	account := entity.AccountDb{
		Id:        991,
		Mail:      "account@test.ru",
		Login:     "sub_account",
		Followers: 0,
	}
	follower := entity.AccountDb{
		Id:    992,
		Mail:  "follower@test.ru",
		Login: "sub_follower",
	}
	storage.CreateAccount(account)
	storage.CreateAccount(follower)
	defer func() {
		storage.db.Where("id = ?", account.Id).Delete(&entity.AccountDb{})
		storage.db.Where("id = ?", follower.Id).Delete(&entity.AccountDb{})
	}()

	err := storage.Subscribe(account.Id, follower.Id)
	if err != nil {
		t.Errorf("ошибка подписки: %v", err)
		return
	}

	var sub entity.Subscribers
	result := storage.db.Where("account_id = ? AND subscriber_id = ?", account.Id, follower.Id).First(&sub)
	if result.Error != nil {
		t.Errorf("запись подписки не найдена: %v", result.Error)
		return
	}

	updated_account, err := storage.GetAccount(account.Id)
	if err != nil {
		t.Errorf("ошибка получения аккаунта: %v", err)
		return
	}
	if updated_account.Followers != account.Followers+1 {
		t.Errorf("Followers не совпадает: got %d, want %d", updated_account.Followers, account.Followers+1)
		return
	}
	t.Log("func [Subscribe] is OK")

	err = storage.Unsubscribe(account.Id, follower.Id)
	if err != nil {
		t.Errorf("ошибка отписки: %v", err)
		return
	}

	result = storage.db.Where("account_id = ? AND subscriber_id = ?", account.Id, follower.Id).First(&sub)
	if result.Error == nil {
		t.Errorf("запись подписки должна быть удалена")
		return
	}

	updated_account, err = storage.GetAccount(account.Id)
	if err != nil {
		t.Errorf("ошибка получения аккаунта: %v", err)
		return
	}
	if updated_account.Followers != account.Followers {
		t.Errorf("Followers не совпадает после отписки: got %d, want %d", updated_account.Followers, account.Followers)
		return
	}
	t.Log("func [Unsubscribe] is OK")
}
