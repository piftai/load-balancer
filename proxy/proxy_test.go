package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReverseProxy(t *testing.T) {
	// Создаем тестовый сервер, который будет имитировать целевой сервер
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer targetServer.Close()

	// Парсим URL тестового сервера
	targetURL, err := url.Parse(targetServer.URL)
	require.NoError(t, err, "Failed to parse target URL")

	// Создаем прокси
	proxy := NewReverseProxy(targetURL)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через прокси
	proxy.ServeHTTP(w, req)

	// Проверяем результат с помощью assert
	assert.Equal(t, http.StatusOK, w.Code, "Unexpected status code")
	assert.Equal(t, "OK", w.Body.String(), "Unexpected response body")
}
