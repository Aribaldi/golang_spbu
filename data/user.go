package data

import (
	"database/sql"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Uuid     string
	Fname    string
	Lname    string
	Username string
	Email    string
	Password string
}

func SaveData(u *User) (int, error) {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var customer_id int
	err_2 := db.QueryRow(`INSERT INTO public.customer (first_name, last_name, login, password) VALUES ($1, $2, $3, $4) RETURNING customer_id;`, u.Fname, u.Lname, u.Email, u.Password).Scan(&customer_id)
	return customer_id, err_2
}

func UserExists(u *User) bool {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var ps, us string
	q, err := db.Query("SELECT login, password FROM public.customer WHERE login = $1 ", u.Email)
	if err != nil {
		return false
	}
	for q.Next() {
		q.Scan(&us, &ps)
	}
	pw := bcrypt.CompareHashAndPassword([]byte(ps), []byte(u.Password))
	if us == u.Email && pw == nil {
		return true
	}
	return false
}

func EncryptPass(password string) string {
	pass := []byte(password)
	hashpw, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	return string(hashpw)
}
