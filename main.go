package main

import (
	"html/template"
	"net/http"

	"food/data"
	"reflect"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

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
		if data.UserExists(u) {
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
	tmpl, _ := template.ParseFiles("./templates/base.html", "./templates/index.html", "./templates/menus.html")
	username := data.GetUserName(r)
	categs := data.FoodCategs()
	if username != "" {
		err := tmpl.ExecuteTemplate(w, "base", categs)
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
		un := r.FormValue("userName")
		pass := data.EncryptPass(r.FormValue("password"))

		u := &data.User{Fname: f, Lname: l, Email: em, Username: un, Password: pass}
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
	http.Handle("/", router)
	CategMenuWrapper()
	http.ListenAndServe(":8000", nil)
}
