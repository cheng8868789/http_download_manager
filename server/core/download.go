package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const basePath = "./../download"

// 请求单个线程所需的数据
func singleThreadDownload(thread *fileProgress, wg *sync.WaitGroup, mu *sync.Mutex, ctx context.Context) {
	defer wg.Done()

	client := &http.Client{}
	req, err := http.NewRequest("GET", thread.URL, nil)
	if err != nil {
		fmt.Println("Error:", err)
		fl.get(thread.FileName).update(2)
		return
	}
	// 设置请求头并获取数据
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", thread.StartPos, thread.EndPos))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		fl.get(thread.FileName).update(2)
		return
	}
	defer resp.Body.Close()
	// 创建文件，并将数据复制到文件中
	file, err := os.OpenFile(thread.FilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error:", err)
		fl.get(thread.FileName).update(2)
		return
	}
	defer file.Close()
	var total int64
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			fl.get(thread.FileName).delete()
			fmt.Println("singleThreadDownload Context done:", ctx.Err())
			return
		default:
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println("Error:", err)
				fl.get(thread.FileName).update(2)
				return
			}
			if n == 0 {
				fl.get(thread.FileName).delete()
				break
			}
			mu.Lock()
			_, err = file.WriteAt(buf[:n], thread.StartPos+total)
			if err != nil {
				mu.Unlock()
				fmt.Println("Error:", err)
				fl.get(thread.FileName).update(2)
				return
			}
			total += int64(n)
			thread.Percentage = float64(total) / float64(thread.EndPos-thread.StartPos+1)
			thread.ProgressInt = int(thread.Percentage * 100)
			mu.Unlock()
		}
	}
}

// 主要逻辑
func DownLoad(url string) {
	// 获取文件名称
	baseName := url[strings.LastIndex(url, "/")+1:]
	fileName := baseName
	existsCount := 0
	// 判断文件是否已经存在,如果存在则在名字后面加序号
	for fl.exists(fileName) {
		existsCount++
		fileName = baseName + " (" + strconv.Itoa(existsCount) + ")"
	}

	// 文件存储路劲
	path := basePath + "/" + fileName
	// 创建文件对象
	file := &fileModel{
		FileName:    fileName,
		URL:         url,
		FilePath:    path,
		Status:      0,
		DeleteCount: 0,
	}
	// 获取文件大小
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 计算每个线程下载的开始和结束位置
	threadNum := coreCount // 线程数
	var startPos, endPos int64
	ForEachSize := fileSize / int64(threadNum)
	downloadThreads := make([]*fileProgress, threadNum)
	for i := 0; i < threadNum; i++ {
		startPos = int64(i) * ForEachSize
		endPos = startPos + ForEachSize - 1
		if i == threadNum-1 {
			endPos = fileSize - 1
		}
		downloadThreads[i] = &fileProgress{
			FileName:    fileName,
			URL:         url,
			FilePath:    path,
			StartPos:    startPos,
			EndPos:      endPos,
			Percentage:  0,
			ProgressInt: 0,
		}
	}
	// 创建context,已便下载异常时关闭所有子协程
	ctx, cancel := context.WithCancel(context.Background())
	// 创建 WaitGroup，开启多个 goroutine 下载
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go singleThreadDownload(downloadThreads[i], &wg, &mu, ctx)
	}
	file.Progress = downloadThreads
	fc.add(file.FileName, cancel)
	// 存入文件队列
	fl.add(file)
	wg.Add(1)
	go fileSuccess(file, &wg, ctx)
	wg.Wait()

}

func fileSuccess(file *fileModel, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	timer := time.NewTicker(1 * time.Second)
	fmt.Println("indeed 1")
	for {
		select {
		case <-ctx.Done():
			file.delete()
			fmt.Println("fileSuccess Context done:", ctx.Err())
			return
		case <-timer.C:
			file.RLock()
			defer file.RUnlock()
			if file.Status != 2 {
				fmt.Println("indeed 2")
				if file.Progress != nil && len(file.Progress) > 0 {
					clearFlag := true
					for _, fileProgress := range file.Progress {
						if fileProgress.ProgressInt != 100 {
							clearFlag = false
						}
					}
					if clearFlag {
						fmt.Println("indeed 3")
						file.Progress = nil
						file.Status = 1
						fmt.Println(file.FileName + " download success")
						fc.delete(file.FileName)
						return
					}
				}
			} else {
				fmt.Println("indeed 4")
				fc.delete(file.FileName)
				return
			}
		}
	}
}
