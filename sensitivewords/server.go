package sensitivewords

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type SWFServer struct {
	HttpServer *http.Server
	WordsTree  *Tree
	ErrorLog   *SWFLogger
	LogCancel  context.CancelFunc
}

// func NewSWFServer(dictPath string) (*SWFServer, error) {
// 	wordsList, err := readDict(dictPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	errLogger, err := NewSWFLogger("error", "4k", 3)
// 	if err != nil {
// 		return nil, err
// 	}
// 	cancel, err := errLogger.Start()
// 	if err != nil {
// 		return nil, err
// 	}
// 	tree := NewTree(wordsList)
// 	mux := http.NewServeMux()
// 	server := &SWFServer{
// 		HttpServer: &http.Server{
// 			Addr:    "localhost:8000",
// 			Handler: mux,
// 		},
// 		WordsTree: tree,
// 		ErrorLog:  errLogger,
// 		LogCancel: cancel,
// 	}
// 	mux.HandleFunc("/sensitive_filter/", server.sensitiveFilter)
// 	return server, nil
// }

func NewSWFServer(args *Args) (*SWFServer, error) {
	wordsList, err := readDict(args.DictPath)
	if err != nil {
		return nil, err
	}
	errLogger, err := NewSWFLogger("error", args.LogSizeLimit, args.LogKeepNumber)
	if err != nil {
		return nil, err
	}
	cancel, err := errLogger.Start()
	if err != nil {
		return nil, err
	}
	tree := NewTree(wordsList)
	mux := http.NewServeMux()
	server := &SWFServer{
		HttpServer: &http.Server{
			Addr:    args.Addr,
			Handler: mux,
		},
		WordsTree: tree,
		ErrorLog:  errLogger,
		LogCancel: cancel,
	}
	mux.HandleFunc("/sensitive_filter/", server.sensitiveFilter)
	return server, nil
}

// func (server *SWFServer) Run() {
// 	server.HttpServer.ListenAndServe()
// }
func (server *SWFServer) Run() {
	go func() {
		if err := server.HttpServer.ListenAndServe(); err != nil {
			server.ErrorLog.Println(err)
		}
	}()
}

type Response struct {
	Status      int      `json:"status"`
	ErrorDetail string   `json:"errorDetail"`
	Data        []string `json:"data"`
}

type Request struct {
	Action string `json:"action"`
	Text   string `json:"text"`
}

func (server *SWFServer) sensitiveFilter(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	response := Response{Data: make([]string, 0)}
	if r.Method != http.MethodPost {
		server.ErrorLog.Println("Method error: from ", r.RemoteAddr)
		response.Status, response.ErrorDetail = 400, "method error"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	var request Request
	err := json.Unmarshal(body, &request)
	if err != nil {
		server.ErrorLog.Println("JSON unmarshal error: from ", r.RemoteAddr, " ", err.Error())
		response.Status, response.ErrorDetail = 401, "JSON unmarshal error"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	if request.Text == "" {
		server.ErrorLog.Println("Empty input error: from ", r.RemoteAddr)
		response.Status, response.ErrorDetail = 402, "empty input"
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
		return
	}
	sensitiveWords := server.WordsTree.search(request.Text)
	response.Status, response.Data = 200, sensitiveWords
	jsonData, _ := json.Marshal(response)
	w.Write(jsonData)
	return
}

func (server *SWFServer) Close() error {
	defer server.LogCancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.HttpServer.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
