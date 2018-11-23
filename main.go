// 面向对象
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

const (
	ID            = 3
	DELETE        = false
	URL_PROJECT   = "https://hub.namifunds.com/api/repositories?project_id=%d&q=&detail=0"
	URL_REGISTORY = "https://hub.namifunds.com/api/repositories/%s/tags?detail=1"
	URL_DELETE    = "https://admin:Harbor@hub.namifunds.com/api/repositories/%s/tags/%s"
)

var now = time.Now()
var countSuccess = 0
var countFailed = 0

type Project struct {
	Name         string
	ID           int
	Repositories []Repository
}

type Repository struct {
	Name string
	Tags []Tag
}

type Tag struct {
	Name       string
	CreateTime string
}

func (p *Project) getRepositories() {
	var repositories []Repository

	url := fmt.Sprintf(URL_PROJECT, p.ID)

	resp, err := getHttpResp(url)
	if err != nil {
		panic("[getHttpResp] error")
	}

	repos := parseRegistory(resp)
	for _, v := range repos {
		// println(v)
		repository := Repository{Name: v}
		repositories = append(repositories, repository)
	}

	p.Repositories = repositories
}

func (r *Repository) getTags() {
	var tags []Tag
	url := fmt.Sprintf(URL_REGISTORY, r.Name)
	resp, err := getHttpResp(url)
	if err != nil {
		panic("[getTags] error")
	}

	for i := 0; ; i++ {
		ii := strconv.Itoa(i)
		tag := Tag{}
		result := gjson.GetBytes(resp, ii)
		if result.Exists() {
			tagName := result.Get("tag").String()

			bigString := result.Get("manifest.history.0.v1Compatibility").String()
			tagCreateTime := GetTagCreateTime(bigString)

			tag.Name = tagName
			tag.CreateTime = tagCreateTime
			tags = append(tags, tag)
		} else {
			r.Tags = tags
			return
		}
	}
}

func main() {
	p := Project{Name: "namibank", ID: ID}
	p.getRepositories()
	// fmt.Println(p.Repositories[0].Name)
	for _, r := range p.Repositories {
		fmt.Printf("【%s】\n", r.Name)
		r.getTags()
		for _, tag := range r.Tags {
			fmt.Printf("    tag: %s\tCreateTime: %s\t", tag.Name, tag.CreateTime)
			days := getDay(tag.CreateTime)
			fmt.Printf("距离今天%d天\t", days)
			if days > 30 && DELETE && tag.Name != "master" && tag.Name != "latest" {
				if ok := deleteTag(r.Name, tag.Name); ok {
					fmt.Printf("删除成功！\n")
					countSuccess++
				} else {
					fmt.Printf("删除失败！！！\n")
					countFailed++
				}
			} else {
				fmt.Printf("跳过\n")
			}
		}
	}
	fmt.Printf("\n\n##### 删除成功：%d，删除失败：%d #####\n", countSuccess, countFailed)
}

func deleteTag(projectName, tagName string) bool {
	url := fmt.Sprintf(URL_DELETE, projectName, tagName)

	client := &http.Client{}
	methodDel, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(methodDel)
	if err != nil {
		panic("http DELETE error")
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200

}

func getDay(createTime string) int {
	var days int
	if createTime == "" {
		days = 0
		return days
	}
	cDate, err := time.Parse("2006-01-02", createTime)
	if err != nil {
		fmt.Errorf("time.Parse error : %s", err)
	}
	sub := now.Sub(cDate)
	hours := sub.Hours()
	days = int(hours / 24)

	return days
}

func GetTagCreateTime(bigString string) string {
	var createDateReg = regexp.MustCompile(`201[0-9]-[01][0-9]-[0123][0-9]`)
	createDate := createDateReg.FindString(bigString)
	return createDate
}
func getHttpResp(url string) ([]byte, error) {
	resp, err := http.Get(url)
	respByte, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return respByte, err
}

func parseRegistory(b []byte) []string {
	var s []string
	for i := 0; ; i++ {
		ii := strconv.Itoa(i)
		result := gjson.GetBytes(b, ii)
		if result.Exists() {
			s = append(s, result.String())
		} else {
			return s
		}
	}
}
