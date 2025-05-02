package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"

	// 1. POST /clients - создание клиента
	createData := `{"id": "test_client", "capacity": 100, "rate_per_second": 10}`
	resp, err := http.Post(baseURL+"/clients", "application/json", bytes.NewBufferString(createData))
	if err != nil {
		fmt.Printf("POST /clients error: %v\n", err)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("POST /clients response: %s\n", body)
		resp.Body.Close()
	}

	//// 2. GET /clients/{id} - получение клиента
	//resp, err = http.Get(baseURL + "/clients/test_client")
	//if err != nil {
	//	fmt.Printf("GET /clients error: %v\n", err)
	//} else {
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("GET /clients response: %s\n", body)
	//	resp.Body.Close()
	//}
	//
	//// 3. PUT /clients - обновление клиента
	//updateData := `{"id": "test_client", "capacity": 200, "rate_per_second": 20}`
	//req, _ := http.NewRequest("PUT", baseURL+"/clients", bytes.NewBufferString(updateData))
	//req.Header.Set("Content-Type", "application/json")
	//client := &http.Client{}
	//resp, err = client.Do(req)
	//if err != nil {
	//	fmt.Printf("PUT /clients error: %v\n", err)
	//} else {
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("PUT /clients response: %s\n", body)
	//	resp.Body.Close()
	//}
	//
	//// 4. DELETE /clients/{id} - удаление клиента
	//req, _ = http.NewRequest("DELETE", baseURL+"/clients/test_client", nil)
	//resp, err = client.Do(req)
	//if err != nil {
	//	fmt.Printf("DELETE /clients error: %v\n", err)
	//} else {
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("DELETE /clients response: %s\n", body)
	//	resp.Body.Close()
	//}
	//
	//// 5. GET / - тест балансировщика
	//resp, err = http.Get(baseURL)
	//if err != nil {
	//	fmt.Printf("GET / error: %v\n", err)
	//} else {
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("GET / response: %s\n", body)
	//	resp.Body.Close()
	//}
}
