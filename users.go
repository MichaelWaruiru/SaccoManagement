package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Email    string
	Username string
	Password []byte
}

var (
	users []User
	store = sessions.NewCookieStore([]byte("fleet-mgnt!"))
)

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirmPassword")

		// Check if passwords match
		if password != confirmPassword {
			setFlashMessage(w, r, "Passwords do not match")
			// http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err.Error())
		}

		// Check if username or email is already taken
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&count)
		if err != nil {
			fmt.Println("Error checking username or email existence:", err)
			return
		}
		if count > 0 {
			setFlashMessage(w, r, "Username or email already exists. Choose another username or email.")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Add the new user to the slice (temporary in-memory storage)
		newUser := User{Email: email, Username: username, Password: hashedPassword}
		users = append(users, newUser)

		_, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
		if err != nil {
			fmt.Println("Error inserting user", err)
			return
		}

		fmt.Println("User registered successfully", username)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// This shows flash message when an existing username signs up
	flashMessage := getFlashMessage(w, r)
	data := struct {
		FlashMessage string
	}{
		FlashMessage: flashMessage,
	}
	tmpl.ExecuteTemplate(w, "signup.html", data)
}

var flashMessageKey = "flashMessage"

func setFlashMessage(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := store.Get(r, "sacco-mgmnt")
	session.Values[flashMessageKey] = message
	session.Save(r, w)
}

func getFlashMessage(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "sacco-mgmnt")
	flashMessage, ok := session.Values[flashMessageKey].(string)
	if !ok {
		return ""
	}
	// Clear the flash message
	delete(session.Values, flashMessageKey)
	session.Save(r, w)
	return flashMessage
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var storedPasswordHash []byte
		err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPasswordHash)
		if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword(storedPasswordHash, []byte(password)) != nil {
			// Invalid login credentials
			setFlashMessage(w, r, "Invalid username or password")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Set a session to mark the user as authenticated
		session, _ := store.Get(r, "sacco-mgmnt")
		session.Values["user"] = username
		session.Save(r, w)

		// Successful login
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Display the login page with flash message
	flashMessage := getFlashMessage(w, r)
	data := struct {
		FlashMessage string
	}{
		FlashMessage: flashMessage,
	}

	tmpl.ExecuteTemplate(w, "login.html", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clears the user session
	session, _ := store.Get(r, "sacco-mgmnt")
	session.Values["user"] = nil
	session.Save(r, w)

	// Set cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Redirect to the login page with cache-busting query parameter
	redirectURL := "/login?nocache=" + time.Now().Format(time.RFC3339)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// func sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		session, err := store.Get(r, "sacco-mgmnt")
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		// Check if session is new or user isn't logged in
// 		if session.IsNew || session.Values["user"] == nil {
// 			http.Redirect(w, r, "/login", http.StatusSeeOther)
// 			return
// 		}

// 		// Check if the user has expired
// 		lastActivity, ok := session.Values["lastActivity"].(int64)
// 		if !ok {
// 			// If lastActivity is not set, treat session as expired
// 			http.Redirect(w, r, "/logout", http.StatusSeeOther)
// 			return
// 		}

// 		// Check if session has expired(15 minutes)
// 		if time.Now().Unix()-lastActivity > 15*60 {
// 			http.Redirect(w, r, "/logout", http.StatusSeeOther)
// 			return
// 		}

// 		// Update last activity time
// 		session.Values["lastActivity"] = time.Now().Unix()
// 		session.Save(r, w)

// 		// Call the next handler
// 		next.ServeHTTP(w, r)
// 	})
// }
