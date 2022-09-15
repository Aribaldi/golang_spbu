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
	Count      int
}

type CartRecord struct {
	Id         int32
	Dish_name  string
	Dish_price float32
	Count      int
	Overall    float32
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

func RemoveFromCart(customer_id int, dish_id int) error {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	_, err := db.Exec("WITH rows AS (SELECT public.cart.cart_id FROM public.cart WHERE public.cart.customer_id = $1 AND public.cart.dish_id = $2 LIMIT 1) DELETE FROM public.cart WHERE public.cart.cart_id IN (SELECT * FROM rows);  ", customer_id, dish_id)
	if err != nil {
		panic(err)
	}
	return err
}

func CartInfo(customer_id int) []CartRecord {
	var db, _ = sql.Open("postgres", db_conn)
	var res []CartRecord
	defer db.Close()
	rows, err := db.Query("SELECT menu.dish_id, menu.dish_name, menu.dish_price, COUNT (menu.dish_id) FROM public.cart, public.menu WHERE public.cart.dish_id = public.menu.dish_id AND public.cart.customer_id = $1 GROUP BY menu.dish_id, menu.dish_name, menu.dish_price", customer_id)

	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var temp CartRecord
		err := rows.Scan(&temp.Id, &temp.Dish_name, &temp.Dish_price, &temp.Count)
		temp.Overall = float32(temp.Count) * temp.Dish_price
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, temp)
	}
	return res
}

func GetUrl(str string) string {
	return "categs/" + strings.Replace(strings.ToLower(str), " ", "_", -1)
}

func AddCateg(categ_name string, descr string) error {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	_, err := db.Exec(`INSERT INTO public.food_categs (categ_name, description) VALUES ($1, $2) ;`, categ_name, descr)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func RemoveCateg(categ string) error {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	_, err := db.Exec("DELETE FROM public.food_categs WHERE public.food_categs.categ_name = $1", categ)
	if err != nil {
		log.Fatal(err)
	}
	return err
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
