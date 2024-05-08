package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandleProxy(t *testing.T) {
	file, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(file.Name())

	reqBody := []byte("test data")
	req := httptest.NewRequest("POST", "/", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	handleProxy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидается статус код %d, получено: %d", http.StatusOK, w.Code)
	}

	_, err = os.Stat(file.Name())
	if os.IsNotExist(err) {
		t.Error("Файл не был создан")
	}

	content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("Ошибка чтения содержимого файла: %v", err)
	}
	if string(content) != string(reqBody) {
		t.Errorf("Содержимое файла не совпадает с отправленными данными. Ожидается: %s, Получено: %s", string(reqBody), string(content))
	}
}
