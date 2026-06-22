package repo

import "Etog/internal/domain/entity"

func FromDbUser(account *entity.AccountDb) *entity.Account {
	newAccount := &entity.Account{
		Id:          account.Id,
		Login:       account.Login,
		Avatar:      account.Avatar,
		Official:    account.Official,
		Description: account.Description,
		Rating:      account.Rating,
		Followers:   account.Followers,
	}
	return newAccount
}
