package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"log"

	_ "github.com/lib/pq"
)

type Dish struct {
	Id         int32
	Dish_name  string
	Dish_price float32
	Category   string
	Dish_descr string
	Count      int
}

type Category struct {
	Name           string
	NormalizedName string
	Description    string
}

type CartRecord struct {
	Id         int32
	DishId     int32
	Dish_name  string
	Dish_price float32
	Count      int
	Overall    float32
}

type Order struct {
	Id int32
	//User  User
	CustomerId  int32
	DateCreated time.Time
	Items       []OrderDetail
}

type OrderDetail struct {
	DishId     int32
	Dish       Dish
	Count      int
	TotalPrice float32
}

const db_conn string = "host=localhost port=5432 user=postgres dbname=food_delivery_golang password=postgres sslmode=disable"

func DishTable(category string) []Dish {
	var res []Dish
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()

	rows, err := db.Query("SELECT dish_id, dish_name, dish_price, dish_descr FROM public.menu WHERE dish_category = $1 ", FindCategoryByName(category).Name)
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

	log.Printf("Executing delete from cart user %d, dish %d", customer_id, dish_id)
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
		err := rows.Scan(&temp.DishId, &temp.Dish_name, &temp.Dish_price, &temp.Count)
		temp.Overall = float32(temp.Count) * temp.Dish_price
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, temp)
	}
	return res
}

func NormalizeCategoryName(category_name string) string {
	return strings.Replace(strings.ToLower(category_name), " ", "_", -1)
}

func GetUrl(str string) string {
	return "menu?category=" + NormalizeCategoryName(str)
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

func CreateOrder(customer_id int32, addr string) {
	var cart = CartInfo(int(customer_id))
	var order_items []OrderDetail

	for _, cart_item := range cart {
		order_items = append(order_items, OrderDetail{DishId: cart_item.DishId, Count: cart_item.Count, TotalPrice: cart_item.Overall})
	}

	var order = Order{CustomerId: customer_id, DateCreated: time.Now().UTC(), Items: order_items}

	var db, _ = sql.Open("postgres", db_conn)
	var order_id int
	err := db.QueryRow(`INSERT INTO public.order (customer_id, datetime, address) VALUES ($1, $2, $3) RETURNING public.order.order_id;`, customer_id, order.DateCreated, addr).Scan(&order_id)
	if err != nil {
		fmt.Println(order_id)
		panic(err)
	}

	for _, item := range order.Items {
		_, err := db.Exec(`INSERT INTO public.order_detail (order_id, dish_id, order_quantity, total_price) VALUES ($1, $2, $3, $4) ;`, order_id, item.DishId, item.Count, item.TotalPrice)
		if err != nil {
			panic(err)
		}
	}

	defer db.Close()
}

func GetOrdersForUser(customer_id int32) []Order {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	rows, err := db.Query("SELECT order_id, customer_id, datetime FROM public.order WHERE public.order.customer_id = $1", customer_id)
	if err != nil {
		panic(err)

	}

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.Id, &order.CustomerId, &order.DateCreated)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}

	return orders
}

func GetOrderDetails(order_id int32) []OrderDetail {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	rows, err := db.Query("SELECT dish_id, order_quantity, total_price from order_detail WHERE public.order_detail.order_id = $1", order_id)
	if err != nil {
		panic(err)
	}

	var items []OrderDetail
	for rows.Next() {
		var item OrderDetail
		err := rows.Scan(&item.DishId, &item.Count, &item.TotalPrice)
		if err != nil {
			panic(err)
		}
		items = append(items, item)
	}

	return items
}

func GetDish(dish_id int32) Dish {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var row = db.QueryRow("SELECT dish_id, dish_name, dish_price, dish_category, dish_descr FROM public.menu WHERE public.menu.dish_id = $1", dish_id)

	var result Dish
	var err = row.Scan(&result.Id, &result.Dish_name, &result.Dish_price, &result.Category, &result.Dish_descr)
	if err != nil {
		panic(err)
	}

	return result
}

func OrderHistory(customer_id int32) []Order {
	var orders = GetOrdersForUser(customer_id)

	for order_idx := range orders {
		orders[order_idx].Items = GetOrderDetails(orders[order_idx].Id)

		for item_idx := range orders[order_idx].Items {
			var dish = GetDish(orders[order_idx].Items[item_idx].DishId)
			orders[order_idx].Items[item_idx].Dish = dish
			log.Println("Dish name")
			log.Println(dish.Dish_name)
		}
	}

	return orders
}

func CleanUserCart(customer_id int32) {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	_, err := db.Exec("DELETE FROM public.cart WHERE public.cart.customer_id = $1", customer_id)
	if err != nil {
		panic(err)
	}
}

func ChangeDishPrice(dish_id int, new_price float32) {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	_, err := db.Exec("UPDATE public.menu SET dish_price = $1 WHERE dish_id = $2", new_price, dish_id)
	if err != nil {
		panic(err)
	}

}

// Get Top 5 popular dish names for provided customer
func GetMostPopularDishNamesForUser(customer_id int32) []string {
	const query = `WITH data AS (
    SELECT dish_name, extract(epoch from ((now() at time zone 'utc') - datetime)) / 3600 as weight
    FROM public.order 
    INNER JOIN public.order_detail USING (order_Id)
    INNER JOIN public.menu USING (dish_id)
    WHERE customer_id = $1
)
SELECT dish_name, sum(0.2 * exp(-0.1 * data.weight)) as total_weight
FROM data
GROUP BY dish_name
ORDER BY total_weight DESC
LIMIT 5`

	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	rows, err := db.Query(query, customer_id)
	if err != nil {
		panic(err)
	}

	var dish_names []string
	for rows.Next() {
		var dish_name string
		var rating float32
		err := rows.Scan(&dish_name, &rating)
		if err != nil {
			panic(err)
		}

		dish_names = append(dish_names, dish_name)
	}

	return dish_names
}

func CategDescription(categ string) string {
	return FindCategoryByName(categ).Description
}

func FindCategoryByName(category_name string) Category {
	log.Println("FindCategoryByName", category_name)

	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var row = db.QueryRow(`SELECT categ_name, description FROM public.food_categs WHERE REPLACE(LOWER(public.food_categs.categ_name), ' ', '_') = REPLACE(LOWER($1), ' ', '_')`, category_name)

	result := Category{NormalizedName: NormalizeCategoryName(category_name)}
	var err = row.Scan(&result.Name, &result.Description)
	if err != nil {
		panic(err)
	}

	return result
}

func CheckCustomer(customer_email string) bool {
	var db, _ = sql.Open("postgres", db_conn)
	defer db.Close()
	var ce string
	q, err := db.Query("SELECT login FROM public.customer WHERE login = $1", customer_email)
	if err != nil {
		return false
	}
	for q.Next() {
		q.Scan(&ce)
	}
	return ce == customer_email
}
