package data

import (
	"database/sql"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Fname    string
	Lname    string
	Email    string
	Password string
	Role     string
}

func SaveData(u *User) (int, error) {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var customer_id int
	err_2 := db.QueryRow(`INSERT INTO public.customer (first_name, last_name, login, password) VALUES ($1, $2, $3, $4) RETURNING customer_id;`, u.Fname, u.Lname, u.Email, u.Password).Scan(&customer_id)
	return customer_id, err_2
}

func UserExists(u *User) User {
	var db, _ = sql.Open("postgres", db_conn)
	var ps, us string
	defer db.Close()
	//var ps, us string
	q, err := db.Query("SELECT login, password, first_name, last_name, customer_id, role FROM public.customer WHERE login = $1 ", u.Email)
	if err != nil {
		return User{}
	}
	for q.Next() {
		q.Scan(&us, &ps, &u.Fname, &u.Lname, &u.Id, &u.Role)
	}
	pw := bcrypt.CompareHashAndPassword([]byte(ps), []byte(u.Password))
	if us == u.Email && pw == nil {
		return *u
	}
	return User{}
}

func EncryptPass(password string) string {
	pass := []byte(password)
	hashpw, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	return string(hashpw)
}
