package main

import (
	"database/sql"
	"fmt"

	"log"
	"strings"

	_ "github.com/lib/pq"
)

type User struct {
	Uuid     string
	Fname    string
	Lname    string
	Username string
	Email    string
	Password string
}

// type FoodCateg struct {
// 	CategId     int
// 	CategName   string
// 	Description string
// }

type Dish struct {
	Dish_name  string
	Dish_price float32
	Dish_descr string
}

func DishTable(category string) []Dish {
	var res []Dish
	var db, _ = sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable")
	defer db.Close()
	rows, err := db.Query("SELECT dish_name, dish_price, dish_descr FROM public.menu WHERE dish_category = $1 ", category)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var temp Dish
		err := rows.Scan(&temp.Dish_name, &temp.Dish_price, &temp.Dish_descr)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, temp)

	}
	fmt.Println(res)
	return res
}

func SaveData(u *User) (int, error) {
	var db, _ = sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable")
	defer db.Close()
	var customer_id int
	err_2 := db.QueryRow(`INSERT INTO public.customer (first_name, last_name, login, password) VALUES ($1, $2, $3, $4) RETURNING customer_id;`, u.Fname, u.Lname, u.Email, u.Password).Scan(&customer_id)
	return customer_id, err_2
}

func GetUrl(str string) string {
	return "example/" + strings.Replace(strings.ToLower(str), " ", "_", -1)
}

func FoodCategs() map[string]string {
	res := make(map[string]string)
	var db, _ = sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable")
	defer db.Close()
	rows, err := db.Query("SELECT categ_name FROM public.food_categs  ")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string
		var url string
		err := rows.Scan(&name)
		url = GetUrl(name)
		if err != nil {
			log.Fatal(err)
		}
		res[name] = url
	}
	return res
}

func userExists(u *User) bool {
	var db, _ = sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable")
	defer db.Close()
	var ps, us string
	q, err := db.Query("SELECT login, password FROM public.customer WHERE login = $1 AND password = $2 ", u.Email, u.Password)
	if err != nil {
		return false
	}
	for q.Next() {
		q.Scan(&us, &ps)
	}
	if us == u.Email && ps == u.Password {
		return true
	}
	return false
}
