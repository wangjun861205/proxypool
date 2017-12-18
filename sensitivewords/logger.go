package sensitivewords

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/go-fsnotify/fsnotify"
)

var basePath = "./log/"

type timeStrSlice []string

func (ts timeStrSlice) Len() int {
	return len(ts)
}

func (ts timeStrSlice) Less(i, j int) bool {
	iTime, _ := time.Parse("2006-01-02 15:04:05", ts[i])
	jTime, _ := time.Parse("2006-01-02 15:04:05", ts[j])
	return iTime.Before(jTime)
}

func (ts timeStrSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

type SWFLogger struct {
	File        *os.File
	SizeLimit   int64
	Logger      *log.Logger
	fileManager *fileManager
}

func NewSWFLogger(fileName string, sizeLimit string, keepNumber int) (*SWFLogger, error) {
	fileManager := newFileManager(fileName, keepNumber)
	file, err := fileManager.newFile()
	if err != nil {
		return nil, err
	}
	pattern := regexp.MustCompile(`(\d+)(\w)`)
	group := pattern.FindStringSubmatch(sizeLimit)
	if len(group[1]) == 0 || len(group[2]) == 0 {
		return nil, fmt.Errorf("%s is not valid size format", sizeLimit)
	}
	var size int64
	switch group[2] {
	case "k", "K":
		sizeK, err := strconv.ParseInt(group[1], 10, 64)
		if err != nil {
			return nil, err
		}
		size = sizeK * 1024
	case "m", "M":
		sizeM, err := strconv.ParseInt(group[1], 10, 64)
		if err != nil {
			return nil, err
		}
		size = sizeM * 1024 * 1024
	case "g", "G":
		sizeG, err := strconv.ParseInt(group[1], 10, 64)
		if err != nil {
			return nil, err
		}
		size = sizeG * 1024 * 1024 * 1024
	default:
		return nil, fmt.Errorf("%s is not a valid size unit", group[2])
	}
	logger := log.New(file, "SensitiveWordsFilter:", log.Ldate|log.Ltime)
	return &SWFLogger{
		File:        file,
		SizeLimit:   size,
		Logger:      logger,
		fileManager: fileManager,
	}, nil
}

func (logger *SWFLogger) Start() (context.CancelFunc, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				watcher.Close()
				return
			case ev := <-watcher.Events:
				if ev.Op == fsnotify.Write {
					info, err := logger.File.Stat()
					if err != nil {
						panic(err)
					}
					if info.Size() > logger.SizeLimit {
						newFile, err := logger.fileManager.newFile()
						if err != nil {
							panic(err)
						}
						err = watcher.Remove(logger.File.Name())
						if err != nil {
							panic(err)
						}
						err = watcher.Add(newFile.Name())
						if err != nil {
							panic(err)
						}
						logger.Logger.SetOutput(newFile)
						logger.File.Close()
						logger.File = newFile
					}
				}
			}
		}
	}()
	watcher.Add(logger.File.Name())
	return cancel, nil
}

func (logger *SWFLogger) Println(v ...interface{}) {
	logger.Logger.Println(v)
}

type fileManager struct {
	fileName   string
	keepNumber int
	watcher    *fsnotify.Watcher
}

func newFileManager(fileName string, keepNumber int) *fileManager {
	return &fileManager{
		fileName:   fileName,
		keepNumber: keepNumber,
	}
}

func (fm *fileManager) newFile() (*os.File, error) {
	filePattern := regexp.MustCompile(fmt.Sprintf(`%s (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`, fm.fileName))
	if _, err := os.Stat("./log/"); os.IsNotExist(err) {
		if err = os.Mkdir("./log/", 0777); err != nil {
			return nil, err
		}
	}
	existFiles, err := ioutil.ReadDir("./log/")
	if err != nil {
		return nil, err
	}
	fileNameList := make(timeStrSlice, 0, fm.keepNumber)
	for _, existFile := range existFiles {
		if !existFile.IsDir() && filePattern.MatchString(existFile.Name()) {
			fileNameList = append(fileNameList, filePattern.FindStringSubmatch(existFile.Name())[1])
		}
	}
	if len(fileNameList) >= fm.keepNumber {
		sort.Sort(fileNameList)
		os.Remove(path.Join("./log/", fm.fileName+" "+fileNameList[0]))
	}

	file, err := os.OpenFile(path.Join("./log/", fm.fileName+" "+time.Now().Format("2006-01-02 15:04:05")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return nil, err
	}
	return file, nil
}
