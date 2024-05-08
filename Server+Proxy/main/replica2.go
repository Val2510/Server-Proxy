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

type User2 struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     string   `json:"age"`
	Friends []string `json:"friends"`
}

type UserStorage2 struct {
	users       map[string]*User2
	usersMutex2 sync.Mutex
}

var (
	userStorage2 = UserStorage2{
		users: make(map[string]*User2),
	}
	filePath2 string = "users.json"
)

func main2() {
	if err := userStorage2.loadFromFile2(filePath2); err != nil {
		fmt.Println("Ошибка загрузки данных о пользователях:", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/create", createUserHandler2)
	r.Post("/make_friends", makeFriendsHandler2)
	r.Delete("/user", deleteUserHandler2)
	r.Get("/friends/{userID}", getFriendsHandler2)
	r.Put("/{userID}", updateUserHandler2)

	fmt.Println("Сервер запущен на http://localhost:8081")
	http.ListenAndServe(":8082", r)
}

func (us *UserStorage2) saveToFile2(filePath2 string) error {
	us.usersMutex2.Lock()
	defer us.usersMutex2.Unlock()

	data, err := json.MarshalIndent(us.users, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath2, data, 0644)
}

func (us *UserStorage2) loadFromFile2(filePath2 string) error {
	file, err := os.Open(filePath2)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	us.usersMutex2.Lock()
	defer us.usersMutex2.Unlock()

	return json.Unmarshal(data, &us.users)
}

func createUserHandler2(w http.ResponseWriter, r *http.Request) {
	var newUser User2
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage2.usersMutex2.Lock()
	defer userStorage2.usersMutex2.Unlock()

	newUser.ID = fmt.Sprintf("%d", len(userStorage2.users)+1)
	userStorage2.users[newUser.ID] = &newUser

	if err := userStorage2.saveToFile2(filePath2); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
}

func makeFriendsHandler2(w http.ResponseWriter, r *http.Request) {
	var friendship struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&friendship)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage2.usersMutex2.Lock()
	defer userStorage2.usersMutex2.Unlock()

	sourceUser, ok1 := userStorage2.users[friendship.SourceID]
	targetUser, ok2 := userStorage2.users[friendship.TargetID]
	if !ok1 || !ok2 {
		http.Error(w, "One of the users not found", http.StatusBadRequest)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, friendship.TargetID)
	targetUser.Friends = append(targetUser.Friends, friendship.SourceID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s и %s теперь друзья\n", sourceUser.Name, targetUser.Name)
}

func deleteUserHandler2(w http.ResponseWriter, r *http.Request) {
	var target struct {
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage2.usersMutex2.Lock()
	defer userStorage2.usersMutex2.Unlock()

	user, ok := userStorage2.users[target.TargetID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(userStorage2.users, target.TargetID)

	for _, u := range userStorage2.users {
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

func getFriendsHandler2(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	userStorage2.usersMutex2.Lock()
	defer userStorage2.usersMutex2.Unlock()

	user, ok := userStorage2.users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Friends)
}

func updateUserHandler2(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var updateAge struct {
		NewAge string `json:"new_age"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateAge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userStorage2.usersMutex2.Lock()
	defer userStorage2.usersMutex2.Unlock()

	user, ok := userStorage2.users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Age = updateAge.NewAge

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully\n")
}
