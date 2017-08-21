package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

/**

 */

const (
	projectID   = 3     // 项目ID
	reserveDays = 16    // 分支保留天数
	realDel     = false // 若想删除分支请改为 true
	urlProject  = "https://admin:admin123@localhost/api/repositories?project_id=%d&q=&detail=0"
	urlTag      = "https://admin:admin123@localhost/api/repositories/%s/tags?detail=1"
	urlDel      = "https://admin:admin123@localhost/api/repositories/%s/tags/%s"
)

var count int = 0

func main() {
	// get all projects
	projects := ListProjects(urlProject, projectID)
	for _, p := range projects {
		fmt.Printf("进入 -- > 【%s】 \n", p)
		DeleteTags(p)
	}
	fmt.Printf("总共删除了：%d 个分支", count)
}

func DeleteTags(project string) {
	url := fmt.Sprintf(urlTag, project)
	// println(url)
	resp, _ := http.Get(url)
	text, _ := ioutil.ReadAll(resp.Body)
	// println(string(text))

	for i := 0; ; i++ {
		thisResult := gjson.GetBytes(text, strconv.Itoa(i))
		if thisResult.Exists() {
			// println("entry for loop")
			// tag
			tagResult := thisResult.Get("tag")
			tag := tagResult.String()
			if tag == "master" || tag == "latest" {
				continue
			}
			fmt.Printf("\t[%s]\n", tag)

			// get v1Compatibility , 大字符串，从里面取create time
			bodyResult := thisResult.Get("manifest.history.0.v1Compatibility")
			body := bodyResult.String()
			// fmt.Println("大字符串：", body)

			tagCreateTime := GetTagCreateTime(body)
			if tagCreateTime > reserveDays && realDel {
				if ok := DoDelete(project, tag); ok {
					fmt.Printf("删除成功\n")
					count++
				}
			} else {
				fmt.Printf("跳过\n")
			}
		} else {
			fmt.Println("")
			return
		}
	}
}

func GetTagCreateTime(body string) int {
	var days int
	var createDateReg = regexp.MustCompile(`201[0-9]-[01][0-9]-[0123][0-9]`)

	createDate := createDateReg.FindString(body)
	fmt.Printf("\t\tcreate Time:%s | ", createDate)
	if createDate == "" {
		days = 0
		return days
	}
	cDate, err := time.Parse("2006-01-02", createDate)
	if err != nil {
		fmt.Errorf("time.Parse error : %s", err)
	}
	now := time.Now()
	sub := now.Sub(cDate)
	hours := sub.Hours()
	days = int(hours / 24)
	fmt.Printf("距离现在 %d 天 | ", days)

	return days
}

func DoDelete(project, tag string) bool {
	url := fmt.Sprintf(urlDel, project, tag)
	// println(url)

	client := &http.Client{}
	methodDel, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(methodDel)
	if err != nil {
		log.Fatal("http DELETE error", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func ListProjectTags(project string) []string {
	url := fmt.Sprintf(urlTag, project)
	// fmt.Println("tag url:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("get tag url error", err)
	}

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(text))
	var tags []string
	err = json.Unmarshal(text, &tags)
	if err != nil {
		log.Fatal(err)
	}
	return tags
}

func ListProjects(urlProject string, id int) []string {
	// println(url)
	url := fmt.Sprintf(urlProject, id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("http.Get error", err)
	}

	// defer resp.Body.Close()
	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ReadAll error", err)
	}

	// fmt.Printf("%s", text)
	var projects []string
	err = json.Unmarshal(text, &projects)
	if err != nil {
		log.Fatal(err)
	}

	return projects
}
