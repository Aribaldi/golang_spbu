package main

import (
	"html/template"
	"net/http"
	"strconv"

	"food/data"
	"reflect"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

type M map[string]interface{}

func GetCategMenu(category string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categ_menu := data.DishTable(category)
		tmpl, err := template.New("tmpl").Funcs(templateFuncs).ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html", "./templates/categ_list.html")
		if err != nil {
			panic(err)
		}
		err = tmpl.ExecuteTemplate(w, "base", categ_menu)
		if err != nil {
			panic(err)
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
		switch r.Method {
		case "POST":
			data.AddToCart(user.Id, id)
		case "GET":
			data.RemoveFromCart(user.Id, id)
			http.Redirect(w, r, "/cart", http.StatusFound)
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
	dish := data.CartInfo(u.Id)
	var sum float32 = 0
	for _, d := range dish {
		sum += d.Overall
	}
	tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/cart.html")
	tmpl.ExecuteTemplate(w, "base", M{"Dish": dish, "Sum": sum})
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
	u := &data.User{}
	tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html")
	err := tmpl.ExecuteTemplate(w, "base", u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("uname")
	pass := r.FormValue("password")
	u := &data.User{Email: name, Password: pass}

	redirect := "/"
	if name != "" && pass != "" {
		if data.UserExists(u) != (data.User{}) {
			data.SetSession(u, w)
			redirect = "/categs"
		}

	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	data.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func categs(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/main.html", "./templates/menus.html")
	user := data.GetUserName(r)
	categs := data.FoodCategs()
	if user != (data.User{}) {
		err := tmpl.ExecuteTemplate(w, "base", M{"categs": categs, "user": user})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl, _ := template.ParseFiles("./templates/signup.html", "./templates/index.html", "./templates/base.html")
		u := &data.User{}
		tmpl.ExecuteTemplate(w, "base", u)
	case "POST":
		f := r.FormValue("fName")
		l := r.FormValue("lName")
		em := r.FormValue("email")
		pass := data.EncryptPass(r.FormValue("password"))

		u := &data.User{Fname: f, Lname: l, Email: em, Password: pass}
		data.SaveData(u)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./templates/static"))))
	router.HandleFunc("/", indexPage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/categs", categs)
	router.HandleFunc("/signup", signup).Methods("POST", "GET")
	router.HandleFunc("/cart", Cart)
	http.Handle("/", router)
	CategMenuWrapper()
	DishWrapper()
	http.ListenAndServe(":8000", nil)
}
