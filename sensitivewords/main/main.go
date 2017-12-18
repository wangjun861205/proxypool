package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"notbear/sensitivewords"
)

func main() {
	addr := flag.String("H", "localhost:8000", "define listen port")
	dictPath := flag.String("D", "./dict/", "define dict path")
	logSizeLimit := flag.String("S", "100m", "define log file size limition")
	logKeepNumber := flag.Int("N", 5, "define log file number limition")
	flag.Parse()
	args := &sensitivewords.Args{
		Addr:          *addr,
		DictPath:      *dictPath,
		LogSizeLimit:  *logSizeLimit,
		LogKeepNumber: *logKeepNumber,
	}
	// sigChan := make(chan os.Signal)
	// signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	// server, err := sensitivewords.NewSWFServer("../originwords/sw1.txt")
	server, err := sensitivewords.NewSWFServer(args)
	if err != nil {
		fmt.Println(err)
		return
	}
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// go server.Run()
	server.Run()
	done := sensitivewords.ListenSignal()
	<-done
	fmt.Println("Closing server...")
	if err = server.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server has closed")
}
