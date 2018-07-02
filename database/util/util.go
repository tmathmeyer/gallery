package util

import (
	"database/sql"
	"../generated"
	"../../ezcrypt"
	"unicode"
	"strings"
	"log"
)

func MakeFriendlyPath(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, str)
}

func GetMetadataValue(db *sql.DB, key string) string {
	data, err := generated.QueryMetadataTable(db, map[string]interface{}{
		"Key": key,
	})
	if err == nil && len(data) > 0 {
		return data[0].Value
	}
	return ""
}

func AddMetadata(db *sql.DB, key string, value string) {
	var metadata generated.Metadata
	metadata.Key = key
	metadata.Value = value

	generated.InsertMetadataTable(db, metadata)
}

func AddAdmin(db *sql.DB, user string, pass string) {
	var admin generated.User
	admin.Name = user
	admin.Admin = 1
	admin.Passhash, _ = ezcrypt.HashPassword(pass)

	generated.InsertUserTable(db, admin)
}

func AddUser(db *sql.DB, user string, pass string) {
	var acct generated.User
	acct.Name = user
	acct.Admin = 0
	acct.Passhash, _ = ezcrypt.HashPassword(pass)
	generated.InsertUserTable(db, acct)
}

func DeleteUser(db *sql.DB, user string) bool {
	if IsUserAdmin(db, user) {
		return false
	}
	return generated.DeleteFromUserTable(db, map[string]interface{}{
		"Name": user,
	}) == nil
}

func ChangePassword(db *sql.DB, user string, newpass string) bool {
	hash, _ := ezcrypt.HashPassword(newpass)
	return generated.UpdateUserTable(db, "Passhash", hash, map[string]interface{}{
		"Name": user,
	}) == nil
}

func ChangePasswordForId(db *sql.DB, id string, newpass string) bool {
	hash, _ := ezcrypt.HashPassword(newpass)
	return generated.UpdateUserTable(db, "Passhash", hash, map[string]interface{}{
		"Id": id,
	}) == nil
}

func ValidCredentials(db *sql.DB, user string, pass string) bool {
	users, err := generated.QueryUserTable(db, map[string]interface{}{
		"Name": user,
	})

	if err != nil || len(users) != 1 {
		return false
	}
	
	return ezcrypt.CheckPasswordHash(pass, users[0].Passhash)
}

func IsUserAdmin(db *sql.DB, user string) bool {
	users, err := generated.QueryUserTable(db, map[string]interface{}{
		"Name": user,
	})

	if err != nil || len(users) != 1 {
		return false
	}

	return users[0].Admin == 1
}