package media

import (
	"bufio"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var medias = make(map[string][]string)

func init() {
	ReloadMedias()
}

func loadFile(f string) []string {
	log.Println("Loading resource:", f)
	file, err := os.Open(f)
	if err != nil {
		log.Println("url file not found")
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	medias := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "mp4") || strings.HasSuffix(line, "gif") {
			continue
		}
		if line == "" {
			continue
		}
		medias = append(medias, line)
	}
	log.Println(f, "contains", len(medias), "items")
	return medias
}

func ReloadMedias() error {
	var files []string
	if err := filepath.Walk("resources", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, f := range files {
		_, filename := filepath.Split(f)
		log.Println(filepath.Split(f))
		medias[strings.TrimSuffix(filename, filepath.Ext(filename))] = loadFile(f)
	}
	log.Println(GetKeywords())
	return nil
}

func GetKeywords() []string {
	keywords := make([]string, 0)
	for k := range medias {
		keywords = append(keywords, k)
	}
	sort.Strings(keywords)
	return keywords
}

func RandomMediaUrl(keyword string, count int) []string {
	urls, exists := medias[keyword]
	if !exists {
		return []string{}
	}
	result := make([]string, 0)
	for i := 0; i < count; i++ {
		result = append(result, urls[rand.Intn(len(urls)-1)])
	}
	return result
}

func DownloadMediaUrl(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 302 {
		return DownloadMediaUrl(resp.Header.Get("Location"))
	}
	if resp.StatusCode == 200 {
		return ioutil.ReadAll(resp.Body)
	}
	return nil, errors.New("download fail with unknown error")
}
