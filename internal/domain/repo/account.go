package repo

import "Etog/internal/domain/entity"

func FromDbUser(account *entity.AccountDb) *entity.Account {
	newAccount := &entity.Account{
		Login:       account.Login,
		Avatar:      account.Avatar,
		Official:    account.Official,
		Description: account.Description,
		Rating:      account.Rating,
		Followers:   account.Followers,
	}
	return newAccount
}
