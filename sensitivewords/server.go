package sensitivewords

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/sensitive_filter/", sensitiveFilter)
	return &http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: mux,
	}
}

type Response struct {
	Status      int      `json:"status"`
	ErrorDetail string   `json:"errorDetail"`
	Data        []string `json:"data"`
}

type Request struct {
	Text string `json:"text"`
}

func sensitiveFilter(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	response := Response{Data: make([]string, 0)}
	if r.Method != http.MethodPost {
		response.Status, response.ErrorDetail = 400, "method error"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	err := r.ParseForm()
	if err != nil {
		response.Status, response.ErrorDetail = 500, "parse post form error"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	var request Request
	err = json.Unmarshal(body, &request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if request.Text == "" {
		response.Status, response.ErrorDetail = 401, "empty input"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	sensitiveWords := Search(request.Text)
	response.Status, response.Data = 200, sensitiveWords
	jsonData, _ := json.Marshal(response)
	w.Write(jsonData)
	return
}
