package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"food/data"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer, "isAdmin": func(user data.User) bool {
	return user.Role == "admin"
}}

type M map[string]interface{}

func GetCategMenu(category string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := data.GetUserName(r)
		if user.Fname != "" {
			switch r.Method {
			case "GET":
				categ_menu := data.DishTable(category)
				tmpl, err := template.New("tmpl").Funcs(templateFuncs).ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html", "./templates/categ_list.html")
				if err != nil {
					panic(err)
				}
				err = tmpl.ExecuteTemplate(w, "base", M{"categ_menu": categ_menu, "user": user})
				if err != nil {
					panic(err)
				}
			case "POST":
				data.RemoveCateg(category)
				http.Redirect(w, r, "/categs", http.StatusFound)
			}
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	})
}

func CategMenuWrapper() {
	categs := data.FoodCategs()
	for k, v := range categs {
		full_path := "/" + v
		http.Handle(full_path, GetCategMenu(k))
	}
}

func AddDish(id int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := data.GetUserName(r)
		if user.Fname != "" {
			switch r.Method {
			case "POST":
				if user.Role == "admin" {
					new_price, _ := strconv.Atoi(r.FormValue("price"))
					data.ChangeDishPrice(id, float32(new_price))
					http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
				} else {
					data.AddToCart(user.Id, id)
					http.Redirect(w, r, "/cart", http.StatusFound)
				}
			case "GET":
				data.RemoveFromCart(user.Id, id)
				http.Redirect(w, r, "/cart", http.StatusFound)
			}
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

	})
}

func DishWrapper() {
	ids := data.DishIds()
	for id := range ids {
		full_path := "/" + "dish/" + strconv.Itoa(id)
		http.Handle(full_path, AddDish(id))
	}
}

func Cart(w http.ResponseWriter, r *http.Request) {
	u := data.GetUserName(r)
	if u.Fname != "" {
		dish := data.CartInfo(u.Id)
		var sum float32 = 0
		for _, d := range dish {
			sum += d.Overall
		}
		tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/cart.html")
		tmpl.ExecuteTemplate(w, "base", M{"Dish": dish, "Sum": sum})
	} else {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func RangeStructer(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return nil
	}
	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Struct {
		return nil
	}

	out := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[i] = v.Field(i).Interface()
	}

	return out
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	msg := data.GetMsg(w, r, "message")
	u := &data.User{}
	u.Errors = make(map[string]string)
	if msg != "" {
		u.Errors["message"] = msg
		tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html")
		err := tmpl.ExecuteTemplate(w, "base", u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		u := &data.User{}
		tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html")
		err := tmpl.ExecuteTemplate(w, "base", u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("uname")
	pass := r.FormValue("password")
	u := &data.User{Email: name, Password: pass}

	redirect := "/"
	if name != "" && pass != "" {
		if data.UserExists(u).Fname != "" {
			data.SetSession(u, w)
			redirect = "/categs"
		} else {
			data.SetMsg(w, "message", "Пожалуйста, зарегестрируйтесь иди введите корректные почту и пароль!")
		}
	} else {
		data.SetMsg(w, "message", "Поле почты или пароля пустые!")
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	data.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func categs(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{"isAdmin": func(user data.User) bool {
		return user.Role == "admin"
	},
	}).ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html", "./templates/menus.html")
	if err != nil {
		panic(err)
	}

	user := data.GetUserName(r)
	categs := data.FoodCategs()
	top_dishes := data.GetMostPopularDishNamesForUser(int32(user.Id))
	if user.Fname != "" {
		err := tmpl.ExecuteTemplate(w, "base", M{"categs": categs, "user": user, "top_dishes": top_dishes})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func add_history_record(w http.ResponseWriter, r *http.Request) {
	user := data.GetUserName(r)
	if user.Fname != "" {
		switch r.Method {
		case "GET":
			tmpl, err := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html", "./templates/addr_confirm.html")
			if err != nil {
				panic(err)
			}
			err2 := tmpl.ExecuteTemplate(w, "base", nil)
			if err2 != nil {
				panic(err2)
			}
		case "POST":
			addr := r.FormValue("addr")
			data.CreateOrder(int32(user.Id), addr)
			data.CleanUserCart(int32(user.Id))
			http.Redirect(w, r, "/final", http.StatusFound)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl, _ := template.ParseFiles("./templates/signup.html", "./templates/index.html", "./templates/base.html")
		u := &data.User{}
		u.Errors = make(map[string]string)
		u.Errors["lname"] = data.GetMsg(w, r, "lname")
		u.Errors["fname"] = data.GetMsg(w, r, "fname")
		u.Errors["email"] = data.GetMsg(w, r, "email")
		u.Errors["username"] = data.GetMsg(w, r, "username")
		u.Errors["password"] = data.GetMsg(w, r, "password")
		tmpl.ExecuteTemplate(w, "base", u)
	case "POST":
		f := r.FormValue("fName")
		l := r.FormValue("lName")
		em := r.FormValue("email")
		pass := r.FormValue("password")
		u := &data.User{Fname: f, Lname: l, Email: em, Password: pass}
		result, err := govalidator.ValidateStruct(u)

		if err != nil {
			e := err.Error()
			if re := strings.Contains(e, "Lname"); re {
				data.SetMsg(w, "lname", "Пожалуйста, введите корректную фамилию!")
			}
			if re := strings.Contains(e, "Email"); re {
				data.SetMsg(w, "email", "Пожалуйста, введите корректный почтовый адрес!")
			}
			if re := strings.Contains(e, "Fname"); re {
				data.SetMsg(w, "fname", "Пожалуйста, введите корректную фамилию!")
			}
			if re := strings.Contains(e, "Password"); re {
				data.SetMsg(w, "password", "Пожалуйста, введите пароль!")
			}

		}

		if r.FormValue("password") != r.FormValue("cpassword") {
			data.SetMsg(w, "password", "Пароли не совпадают!")
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		if result {
			u.Password = data.EncryptPass(pass)
			data.SaveData(u)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/signup", http.StatusFound)

	}
}

func AddCategForm(w http.ResponseWriter, r *http.Request) {
	u := &data.User{}
	if u.Role == "admin" {
		tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/add_categ.html")
		err := tmpl.ExecuteTemplate(w, "base", u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

func AddCateg(w http.ResponseWriter, r *http.Request) {
	u := &data.User{}
	if u.Role == "admin" {
		categ_name := r.FormValue("categ")
		categ_descr := r.FormValue("description")
		data.AddCateg(categ_name, categ_descr)
		http.Redirect(w, r, "/categs", http.StatusFound)

	} else {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}

}

func ViewOrdersHistory(w http.ResponseWriter, r *http.Request) {
	user := data.GetUserName(r)
	if user.Fname != "" {
		orders := data.OrderHistory(int32(user.Id))
		log.Println("Showing orders for user", user.Id)
		log.Println(orders)
		tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/orders_hist.html")
		err := tmpl.ExecuteTemplate(w, "base", M{"orders": orders, "user": user})
		if err != nil {
			panic(err)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

}

func final(w http.ResponseWriter, r *http.Request) {
	tmpl, err1 := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/final.html")
	if err1 != nil {
		panic(err1)
	}
	err := tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./templates/static"))))
	router.HandleFunc("/", indexPage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/add_categ_form", AddCategForm)
	router.HandleFunc("/add_categ", AddCateg).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/categs", categs)
	router.HandleFunc("/signup", signup).Methods("POST", "GET")
	router.HandleFunc("/addr_form", add_history_record).Methods("POST", "GET")
	router.HandleFunc("/cart", Cart)
	router.HandleFunc("/history", ViewOrdersHistory)
	router.HandleFunc("/final", final)
	http.Handle("/", router)
	CategMenuWrapper()
	DishWrapper()
	http.ListenAndServe(":8000", nil)
}
