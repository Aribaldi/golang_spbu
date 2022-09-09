package main

import (
	"fmt"
	"html/template"
	"net/http"

	"reflect"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var router = mux.NewRouter()

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

// var htmlTemplate = `{{range .}}<tr>
// {{range rangeStruct .}} <td>{{.}}</td>
// {{end}}</tr>
// {{end}}`

func GetCategMenu(category string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		categ_menu := DishTable(category)
		tmpl, err := template.New("tmpl").Funcs(templateFuncs).ParseFiles("./user/base.html", "./user/index.html", "./user/main.html", "./user/categ_list.html")
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
	categs := FoodCategs()
	for k, v := range categs {
		full_path := "/" + v
		fmt.Println(full_path)
		http.Handle(full_path, GetCategMenu(k))
	}
}

func RangeStructer(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return nil
	}
	fmt.Println(args[0])
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
	u := &User{}
	tmpl, _ := template.ParseFiles("./user/base.html", "./user/index.html", "./user/main.html")
	err := tmpl.ExecuteTemplate(w, "base", u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("uname")
	pass := r.FormValue("password")
	u := &User{Email: name, Password: pass}

	redirect := "/"
	if name != "" && pass != "" {
		if userExists(u) {
			setSession(u, w)
			redirect = "/example"
		}

	}
	http.Redirect(w, r, redirect, 302)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", 302)
}

func examplePage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("./user/base.html", "./user/index.html", "./user/internal.html", "./user/menus.html")
	username := getUserName(r)
	categs := FoodCategs()
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
		tmpl, _ := template.ParseFiles("./user/signup.html", "./user/index.html", "./user/base.html")
		u := &User{}
		tmpl.ExecuteTemplate(w, "base", u)
	case "POST":
		f := r.FormValue("fName")
		l := r.FormValue("lName")
		em := r.FormValue("email")
		un := r.FormValue("userName")
		pass := r.FormValue("password")

		u := &User{Fname: f, Lname: l, Email: em, Username: un, Password: pass}
		SaveData(u)
		http.Redirect(w, r, "/", 302)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", indexPage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/example", examplePage)
	router.HandleFunc("/signup", signup).Methods("POST", "GET")
	http.Handle("/", router)
	CategMenuWrapper()
	http.ListenAndServe(":8000", nil)
}
