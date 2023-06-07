package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"

	_ "embed"

	"github.com/beevik/etree"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
)

//go:embed config.json
var defaultConfigContent []byte

func Translate(source, targetLang string) (string, error) {
	encodedSource := url.QueryEscape(source)

	url := "https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&dj=1&ie=UTF-8&sl=auto&tl=" + targetLang + "&q=" + encodedSource

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Join(errors.New("Error getting translate.googleapis.com"))
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Error getting translate.googleapis.com")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	results := gjson.GetBytes(body, "sentences.#.trans")
	if !results.Exists() {
		return "", errors.New("Json data returned from translate.googleapis.com is invalid")
	}

	result := lo.Reduce(results.Array(), func(agg string, item gjson.Result, index int) string {
		return agg + item.String()
	}, "")

	return result, nil
}

func Init(c *cli.Context) error {
	return ioutil.WriteFile("./config.json", defaultConfigContent, 0644)
}

func Run(c *cli.Context) error {
	var (
		err error
	)

	configContentBytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	configContent := string(configContentBytes)

	// 使用gjson读取文件
	host := gjson.Get(configContent, "host").String()
	port := gjson.Get(configContent, "port").String()
	language := gjson.Get(configContent, "language").String()
	interval := gjson.Get(configContent, "cron").String()
	rss := gjson.Get(configContent, "rss").Array()

	// 创建path与content对应的map
	pathContentMap := make(map[string]string)
	// 创建path与contentType对应的map
	pathContentTypeMap := make(map[string]string)
	// 创建path与fun对应的map
	pathFunMap := make(map[string]func())

	// 创建定时任务
	crontab := cron.New(cron.WithSeconds())
	fmt.Printf("[Info] refresh crontab: %s\n", interval)
	for _, r := range rss {
		url := r.Get("url").String()
		path := r.Get("path").String()

		fun := func() {
			r := r
			now := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("[%s] [Info] %s: refresh %s\n", now, path, url)
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			pathContentTypeMap[path] = resp.Header.Get("Content-Type")

			doc := etree.NewDocument()
			_, err = doc.ReadFrom(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			for _, item := range doc.FindElements(r.Get("xml_item_path").String()) {
				title := item.SelectElement(r.Get("xml_title_in_item_path").String())
				text := title.Text()
				result, err := Translate(text, language)
				if err != nil {
					continue
				}
				title.SetCData(fmt.Sprintf("%s<br/>%s", result, text))
			}
			doc.Indent(2)
			content, err := doc.WriteToString()
			if err != nil {
				log.Fatal(err)
			}
			pathContentMap[path] = content
		}
		pathFunMap[r.Get("path").String()] = fun
		// 立即执行一次
		fun()

	}
	crontab.AddFunc(interval, func() {
		for _, fun := range pathFunMap {
			fun()
		}
	})
	crontab.Start()
	defer crontab.Stop()

	gin.SetMode(gin.ReleaseMode)
	// 设置gin路由
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	for path := range pathContentMap {
		p := path
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		r.GET(p, func(c *gin.Context) {
			c.Header("Content-Type", pathContentTypeMap[p])
			c.String(http.StatusOK, pathContentMap[p])
		})
	}

	return r.Run(host + ":" + port)
}

func main() {
	app := &cli.App{
		Name:    "RSS-Translator",
		Usage:   "translate rss title with google tansalte",
		Version: "v0.0.1",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "run rss translator",
				Action: Run,
			},
			{
				Name:   "init",
				Usage:  "init rss translator config",
				Action: Init,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
