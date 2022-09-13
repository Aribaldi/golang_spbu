package data

import (
	"database/sql"
	"strings"

	"log"

	_ "github.com/lib/pq"
)

type Dish struct {
	Id         int32
	Dish_name  string
	Dish_price float32
	Dish_descr string
}

const db_conn string = "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable"

func DishTable(category string) []Dish {
	var res []Dish
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	rows, err := db.Query("SELECT dish_id, dish_name, dish_price, dish_descr FROM public.menu WHERE dish_category = $1 ", category)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var temp Dish
		err := rows.Scan(&temp.Id, &temp.Dish_name, &temp.Dish_price, &temp.Dish_descr)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, temp)

	}
	return res
}

func DishIds() []int {
	var res []int
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	rows, err := db.Query("SELECT dish_id FROM public.menu  ")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, id)

	}
	return res
}

func AddToCart(customer_id int, dish_id int) error {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var cart_id int
	err_2 := db.QueryRow(`INSERT INTO public.cart (customer_id, dish_id) VALUES ($1, $2) RETURNING cart_id;`, customer_id, dish_id).Scan(&cart_id)
	if err_2 != nil {
		log.Fatal(err_2)
	}
	return err_2
}

func GetUrl(str string) string {
	return "categs/" + strings.Replace(strings.ToLower(str), " ", "_", -1)
}

func FoodCategs() map[string]string {
	res := make(map[string]string)
	var db, _ = sql.Open("postgres", db_conn)
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
