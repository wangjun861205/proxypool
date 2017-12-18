package sensitivewords

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

func readDict(dictPath string) ([]string, error) {
	wordsList := make([]string, 0, 10240)
	fileInfo, err := os.Stat(dictPath)
	if err != nil {
		return wordsList, err
	}
	if fileInfo.IsDir() {
		fileInfos, err := ioutil.ReadDir(dictPath)
		if err != nil {
			return wordsList, err
		}
		for _, info := range fileInfos {
			f, err := os.Open(path.Join(dictPath, info.Name()))
			if err != nil {
				return wordsList, err
			}
			defer f.Close()
			byteContent, err := ioutil.ReadAll(f)
			if err != nil {
				return wordsList, err
			}
			strContent := string(byteContent[:])
			strContent = strings.Replace(strContent, "\r", "", -1)
			strContent = strings.Replace(strContent, " ", "", -1)
			words := strings.Split(strContent, "\n")
			for _, word := range words {
				wordsList = append(wordsList, word)
			}
		}
		return wordsList, err
	}
	f, err := os.Open(dictPath)
	if err != nil {
		return wordsList, err
	}
	defer f.Close()
	byteContent, err := ioutil.ReadAll(f)
	if err != nil {
		return wordsList, err
	}
	strContent := string(byteContent[:])
	strContent = strings.Replace(strContent, "\r", "", -1)
	strContent = strings.Replace(strContent, " ", "", -1)
	words := strings.Split(strContent, "\n")
	return words, nil
}

func ListenSignal() chan interface{} {
	signalChan := make(chan os.Signal)
	doneChan := make(chan interface{})
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		sig := <-signalChan
		fmt.Printf("\ngot %s signal\n", sig)
		close(doneChan)
		return
	}()
	return doneChan
}

type Args struct {
	Addr          string
	DictPath      string
	LogSizeLimit  string
	LogKeepNumber int
}
