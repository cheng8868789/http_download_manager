package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
)

type fileModel struct {
	FileName string          `json:"fileName"` // 文件名称
	URL      string          `json:"url"`      // 要下载文件的 URL 地址
	FilePath string          `json:"filePath"` // 下载完成后保存的相对路径
	Status   int             `json:"status"`   //0-下载/1-成功/2-失败
	Progress []*fileProgress `json:"progress"`
	sync.RWMutex
}

type fileProgress struct {
	FileName    string // 文件名称
	URL         string // 要下载文件的 URL 地址
	FilePath    string // 下载完成后保存的文件路径
	StartPos    int64  // 线程开始下载的位置
	EndPos      int64  // 线程结束下载的位置
	Percentage  float64
	ProgressInt int `json:"progressInt"`
}

func (f *fileModel) update(status int) {
	f.Lock()
	defer f.Unlock()
	f.Status = status
	return
}

type fileList struct {
	FileListMap map[string]*fileModel `json:"file_list"`
	sync.RWMutex
}

type fileWG struct {
	Wg     *sync.WaitGroup
	Cancel context.CancelFunc
}

type fileWgList struct {
	FileWgList map[string]*fileWG
	sync.Mutex
}

var fl *fileList
var fc *fileWgList

func (fc *fileWgList) add(key string, fileWg *fileWG) {
	fc.Lock()
	defer fc.Unlock()
	fc.FileWgList[key] = fileWg
}

func (fc *fileWgList) delete(key string) {
	fc.Lock()
	defer fc.Unlock()
	cancel, exists := fc.FileWgList[key]
	if !exists {
		return
	}
	cancel.Cancel()
	cancel.Wg.Wait()
	delete(fc.FileWgList, key)
}

func (fl *fileList) add(f *fileModel) {
	fl.Lock()
	defer fl.Unlock()
	fl.FileListMap[f.FileName] = f
}

func (fl *fileList) get(key string) *fileModel {
	fl.RLock()
	defer fl.RUnlock()
	_, exists := fl.FileListMap[key]
	if !exists {
		return nil
	}
	return fl.FileListMap[key]
}

func (fl *fileList) exists(key string) bool {
	fl.RLock()
	defer fl.RUnlock()
	_, exists := fl.FileListMap[key]
	return exists
}

func (fl *fileList) delete(key string) {
	fl.Lock()
	defer fl.Unlock()
	delete(fl.FileListMap, key)
}

// 获取当前机器的核心数
var coreCount int

func Init() error {
	// 32个下载协程
	coreCount = 32
	fl = &fileList{}
	fl.FileListMap = make(map[string]*fileModel)
	fc = &fileWgList{}
	fc.FileWgList = make(map[string]*fileWG)
	//读取download文件夹下的文件目录，初始化文件列表信息
	dir, err := os.Open("./../download")
	if err != nil {
		fmt.Println("打开download目录失败:", err)
		return errors.New("打开download目录失败:" + err.Error())
	}
	defer dir.Close()

	items, err := dir.Readdir(-1)
	if err != nil {
		fmt.Println("读取目录信息失败:", err)
		return errors.New("读取目录信息失败:" + err.Error())
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		} else {
			fl.add(&fileModel{
				FileName: item.Name(),
				Status:   1,
			})
		}
	}
	return nil
}

func DownLoadProcess() []*fileModel {
	fl.RLock()
	defer fl.RUnlock()
	fileList := make([]*fileModel, 0)
	var keys []string
	for key, _ := range fl.FileListMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fileList = append(fileList, fl.FileListMap[key])
	}
	return fileList
}

func StopDownLoad(fileName string) error {
	file := fl.get(fileName)
	if file == nil {
		return errors.New("文件不存在")
	}
	switch file.Status {
	case 1, 2:
		return deleteFile(fileName)
	case 0:
		fc.delete(fileName)
		return deleteFile(fileName)
	default:
		return nil
	}
}

func deleteFile(fileName string) error {
	err := os.Remove(basePath + "/" + fileName)
	if err != nil {
		fmt.Println("文件删除失败:", err.Error())
		return err
	}
	fmt.Println("文件删除成功")
	fl.delete(fileName)
	return nil
}
