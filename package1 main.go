package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

////Структуры

type Note_task struct {
	// Id, User_id, Group_id int
	Id                           int
	Title, Note, Color, Due_date string
}

var posts = []Note_task{}

type ContactDetails struct { //Вход
	Login    string
	Password string
	Success  bool
}

type User_prof struct {
	Id            int
	Login         string
	Hash_password string
	Email         string
}

var data = ContactDetails{}

type Group_task struct {
	Id      int
	User_id int
	// Task_id		  int
	group_name string
	Token      string
}

type PageData struct {
	Success bool
	Posts   []Note_task
}

type usergrouprelation struct {
	Id        int
	User_id   int
	Groupe_id int
}

var group = Group_task{}

///////////////////////////////////////////PAGES////////////////////////////////////////////////////

func index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/index.html", "templates/hider.html", "templates/footer.html", "templates/error.html", "templates/panel.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	//выборка данных
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//выборка данных
	res, err := db.Query("SELECT task_id, task_name, task_description, task_color, due_date FROM `task`")
	if err != nil {
		panic(err)
	}
	posts = []Note_task{}
	for res.Next() {
		var notet Note_task
		err = res.Scan(&notet.Id, &notet.Title, &notet.Note, &notet.Color, &notet.Due_date)
		if err != nil {
			panic(err)
		}
		posts = append(posts, notet)

	}
	fmt.Println("index:", data.Success)

	if data.Success {

		tmpl.ExecuteTemplate(w, "hider", data)

		pageData := PageData{
			Success: data.Success,
			Posts:   posts,
		}

		tmpl.ExecuteTemplate(w, "index", pageData)

	} else {
		tmpl.ExecuteTemplate(w, "hider", data)
		tmpl.ExecuteTemplate(w, "index", data)

	}

}

func create(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/create.html", "templates/hider.html", "templates/footer.html", "templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//выборка данных
	res, err := db.Query("SELECT  * FROM `task`") //продолжить!!!!!!!!!!!!!!!
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(w, "hider", data)
	tmpl.ExecuteTemplate(w, "create", nil)

}

func create_group(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/group.html", "templates/hider.html", "templates/footer.html", "templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	Token, err := generateGroupToken()
	if err != nil {
		log.Fatal(err)
	}

	tmpl.ExecuteTemplate(w, "hider", data)
	tmpl.ExecuteTemplate(w, "create_group", Token)
}

func join_group(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/join_group.html", "templates/hider.html", "templates/footer.html", "templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpl.ExecuteTemplate(w, "hider", data)
	tmpl.ExecuteTemplate(w, "join_group", nil)
}

func registration_page(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/registration_page.html", "templates/hider.html", "templates/footer.html", "templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpl.ExecuteTemplate(w, "registration", nil)
}

func join_page(w http.ResponseWriter, r *http.Request) { //ВХОД

	tmpl, err := template.ParseFiles("templates/join_page.html", "templates/index.html", "templates/hider.html", "templates/footer.html", "templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpl.ExecuteTemplate(w, "join", nil)
}

func error_page(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	tmpl.ExecuteTemplate(w, "error_page", nil)
}

func about_page(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/about_page.html", "templates/hider.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	tmpl.ExecuteTemplate(w, "hider", data)
	tmpl.ExecuteTemplate(w, "about_page", nil)
}

///////////////////////////////FUNCTION//////////////////////////////

func registration_user(w http.ResponseWriter, r *http.Request) { //регистрация

	name := r.FormValue("username")
	e_mail := r.FormValue("email")
	password := r.FormValue("password")

	if password == "" || name == "" || e_mail == "" {
		http.Redirect(w, r, "/error_page/", http.StatusSeeOther)

	} else {

		hashedPassword, err := HashPassword(password)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Hashed Password:", hashedPassword)

		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		registrationDate := time.Now()

		insert, err := db.Query(fmt.Sprintf("INSERT INTO user (username,password,email,registration_date) VALUES ('%s','%s','%s','%s')", name, hashedPassword, e_mail, registrationDate.Format("2006-01-02")))

		if err != nil {
			panic(err)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func save_article_task(w http.ResponseWriter, r *http.Request) { //обработка данных

	title := r.FormValue("title")
	note := r.FormValue("note")
	colorPicker := r.FormValue("colorPicker")
	datePicker := r.FormValue("datePicker")
	groupSelect := r.FormValue("groupSelect")

	user, err := getUserByUsername(data.Login)
	if err != nil {
		log.Fatal(err)
	}

	if title == "" || note == "" || datePicker == "" {
		http.Redirect(w, r, "/error_page/", http.StatusSeeOther)

	} else {
		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		insert, err := db.Query(fmt.Sprintf("INSERT INTO task (task_name, task_description, task_color, due_date, user_id) VALUES ('%s','%s','%s','%s','%d')", title, note, colorPicker, datePicker, user.Id))

		if err != nil {
			panic(err)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func save_group(w http.ResponseWriter, r *http.Request) { //сохранение группы

	user, err := getUserByUsername(data.Login)
	if err != nil {
		log.Fatal(err)
	}

	group = Group_task{
		group_name: r.FormValue("name_group"),
		Token:      r.FormValue("token"),
	}

	if group.group_name == "" || group.group_name == "" {
		http.Redirect(w, r, "/error_page/", http.StatusSeeOther)

	} else {
		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		insert, err := db.Query(fmt.Sprintf("INSERT INTO group_task (group_name, token, created_by_user_id) VALUES ('%s','%s','%d')", group.group_name, group.Token, user.Id))

		if err != nil {
			panic(err)
		}
		defer insert.Close()

		take_groupID, err := getIdGroup(group.Token)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("ID:", take_groupID.Id)

		insert_2, err := db.Query(fmt.Sprintf("INSERT INTO usergrouprelation (user_id , group_id) VALUES ('%d','%d')", user.Id, take_groupID.Id))

		if err != nil {
			panic(err)
		}
		defer insert_2.Close()

		// http.Redirect(w, r, "/save_usergrouprelation/", http.StatusSeeOther)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func save_join_group(w http.ResponseWriter, r *http.Request) { // вступить в группу

	token := r.FormValue("token_group")

	user, err := getUserByUsername(data.Login)
	if err != nil {
		log.Fatal(err)
	}

	if token == "" {
		http.Redirect(w, r, "/error_page/", http.StatusSeeOther)

	} else {
		db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		take_groupID, err := getIdGroup(token)
		if err != nil {
			http.Redirect(w, r, "/error_page/", http.StatusSeeOther)
			log.Fatal(err)
		}
		fmt.Println("ID:", take_groupID.Id)

		insert_2, err := db.Query(fmt.Sprintf("INSERT INTO usergrouprelation (user_id , group_id) VALUES ('%d','%d')", user.Id, take_groupID.Id))

		if err != nil {
			panic(err)
		}
		defer insert_2.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func deleteTask(w http.ResponseWriter, r *http.Request) { //УДАЛЕНИЕ
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id := r.FormValue("id") // Получаем значение id из параметра запроса

	// Выполняем SQL-запрос для удаления записи по указанному id
	_, err = db.Exec("DELETE FROM task WHERE task_id = ?", id)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getIdGroup(grouptoken string) (Group_task, error) {
	var take_grope Group_task

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		return take_grope, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT group_id FROM `group_task` WHERE token = ?", grouptoken)
	err = row.Scan(&take_grope.Id)
	if err != nil {
		return take_grope, err
	}

	return take_grope, nil
}

func getUserByUsername(username string) (User_prof, error) {
	var user User_prof

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		return user, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT user_id FROM `user` WHERE username = ?", username)
	err = row.Scan(&user.Id)
	if err != nil {
		return user, err
	}

	return user, nil
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash проверяет пароль на соответствие хешу
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateGroupToken создает токен для группы
func generateGroupToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func handler(w http.ResponseWriter, r *http.Request) { //ВХОД

	data = ContactDetails{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mainsite")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Query("SELECT user_id, username, password, email FROM `user`")
	if err != nil {
		panic(err)
	}
	data.Success = false
	for res.Next() {

		fmt.Println(data.Success)
		var checkUser User_prof
		err = res.Scan(&checkUser.Id, &checkUser.Login, &checkUser.Hash_password, &checkUser.Email)
		if err != nil {
			panic(err)
		}

		if CheckPasswordHash(data.Password, checkUser.Hash_password) && data.Login == checkUser.Login {
			data.Success = true
			fmt.Println("Success")
			break
		}
	}

	if data.Success {
		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else {
		http.Redirect(w, r, "/error_page/", http.StatusSeeOther)
	}

}

func exit_p(w http.ResponseWriter, r *http.Request) {

	data.Success = false
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//////////////////////////////////////Request/////////////////////////////

func handleRequest() { //ссылки
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)

	http.HandleFunc("/create/", create)
	http.HandleFunc("/create_group/", create_group)
	http.HandleFunc("/join_group/", join_group)
	http.HandleFunc("/join_group_in/", save_join_group)
	http.HandleFunc("/save_article/", save_article_task)
	http.HandleFunc("/save_group/", save_group)
	http.HandleFunc("/delete_task/", deleteTask)

	http.HandleFunc("/registration/", registration_page)
	http.HandleFunc("/login/", join_page)
	http.HandleFunc("/join/", handler)
	http.HandleFunc("/Exit/", exit_p)
	http.HandleFunc("/save_user/", registration_user)
	http.HandleFunc("/error_page/", error_page)
	http.HandleFunc("/about/", about_page)
	http.ListenAndServe(":8000", nil)
}

func main() {
	handleRequest()

}
