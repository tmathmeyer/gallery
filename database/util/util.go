package util

import (
	"database/sql"
	"../generated"
	"../../ezcrypt"
	"unicode"
	"strings"
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