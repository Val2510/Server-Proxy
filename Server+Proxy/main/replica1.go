package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     string   `json:"age"`
	Friends []string `json:"friends"`
}

type UserStorage struct {
	users      map[string]*User
	usersMutex sync.Mutex
}

var (
	userStorage = UserStorage{
		users: make(map[string]*User),
	}
	filePath string = "users.json"
)

func main() {
	if err := userStorage.loadFromFile(filePath); err != nil {
		fmt.Println("Ошибка загрузки данных о пользователях:", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/create", createUserHandler)
	r.Post("/make_friends", makeFriendsHandler)
	r.Delete("/user", deleteUserHandler)
	r.Get("/friends/{userID}", getFriendsHandler)
	r.Put("/{userID}", updateUserHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8081", r)
}

func (us *UserStorage) saveToFile(filePath string) error {
	us.usersMutex.Lock()
	defer us.usersMutex.Unlock()

	data, err := json.MarshalIndent(us.users, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

func (us *UserStorage) loadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	us.usersMutex.Lock()
	defer us.usersMutex.Unlock()

	return json.Unmarshal(data, &us.users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage.usersMutex.Lock()
	defer userStorage.usersMutex.Unlock()

	newUser.ID = fmt.Sprintf("%d", len(userStorage.users)+1)
	userStorage.users[newUser.ID] = &newUser

	if err := userStorage.saveToFile(filePath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
}

func makeFriendsHandler(w http.ResponseWriter, r *http.Request) {
	var friendship struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&friendship)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage.usersMutex.Lock()
	defer userStorage.usersMutex.Unlock()

	sourceUser, ok1 := userStorage.users[friendship.SourceID]
	targetUser, ok2 := userStorage.users[friendship.TargetID]
	if !ok1 || !ok2 {
		http.Error(w, "One of the users not found", http.StatusBadRequest)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, friendship.TargetID)
	targetUser.Friends = append(targetUser.Friends, friendship.SourceID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s и %s теперь друзья\n", sourceUser.Name, targetUser.Name)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var target struct {
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage.usersMutex.Lock()
	defer userStorage.usersMutex.Unlock()

	user, ok := userStorage.users[target.TargetID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(userStorage.users, target.TargetID)

	for _, u := range userStorage.users {
		for i, friendID := range u.Friends {
			if friendID == target.TargetID {
				u.Friends = append(u.Friends[:i], u.Friends[i+1:]...)
				break
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Удалён пользователь: %s\n", user.Name)
}

func getFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	userStorage.usersMutex.Lock()
	defer userStorage.usersMutex.Unlock()

	user, ok := userStorage.users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Friends)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var updateAge struct {
		NewAge string `json:"new_age"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateAge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage.usersMutex.Lock()
	defer userStorage.usersMutex.Unlock()

	user, ok := userStorage.users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Age = updateAge.NewAge

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully\n")
}
