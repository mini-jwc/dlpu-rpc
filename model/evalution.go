package model

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	urlPak "net/url"
	"regexp"
)

func (u *User) Evaluation(ctx context.Context, args *Args, reply *Reply) error {

	var url1 string
	if args.Data == `` {
		url1 = args.Ip + `:8080/jsxsd/xspj/xspj_find.do`
	} else {
		str, _ := urlPak.QueryUnescape(args.Data)
		url1 = args.Ip + `:8080/jsxsd/xspj/xspj_list.do` + str
	}
	_, doc, err := GetWebPage(&url1, args.Session, true)
	if err != nil {
		reply.Err = err
		return nil
	}
	if args.Data == `` {

		if err != nil {
			fmt.Println(err)
		}
		var rep EvaRep1
		sec := doc.Find(`form`).Find(`tr`).Eq(1)
		if sec.Length() != 0 {
			sec1 := sec.Find(`*`)
			rep.Name = sec1.Eq(3).Text()
			rep.StartTime = sec1.Eq(4).Text()
			rep.EndTime = sec1.Eq(5).Text()
			sec1.Find(`a`).Each(func(i int, selection *goquery.Selection) {
				url, err := selection.Attr(`href`)
				if err == false {
					return
				}
				rep.Urls = append(rep.Urls, url[24:])
			})
		}
		reply.Res = rep
		return nil
	} else {
		if err != nil {
			fmt.Println(err)
		}
		var rep []EvaRep2
		sec := doc.Find(`#dataList`).Find(`tr`)
		if sec.Length() != 1 {
			sec.Each(func(i int, selection *goquery.Selection) {
				if i == 0 {
					return
				}
				url, ext := selection.Find(`a`).Attr(`href`)
				if ext == false {
					return
				}
				selection = selection.Find(`*`)
				score := selection.Eq(4).Text()
				if score != "0" && selection.Eq(6).Text() == "否" {
					score = "0"
				}
				rep = append(rep, EvaRep2{
					Name:    selection.Eq(2).Text(),
					Teacher: selection.Eq(3).Text(),
					Score:   score,
					Url:     url[42 : len(url)-11],
				})
			})
		}
		reply.Res = rep
		return nil
	}
}

func (u *User) EvaluationDetail(ctx context.Context, args *Args, reply *Reply) error {
	if args.Data == `` {
		reply.Res = `参数错误`
		return nil
	}
	str, _ := urlPak.QueryUnescape(args.Data)
	url := args.Ip + `:8080/jsxsd/xspj/xspj_edit.do` + str
	_, doc, err := GetWebPage(&url, args.Session, true)
	if err != nil {
		reply.Err = err
		return nil
	}
	var repdata RepForm
	form := doc.Find(`form`)
	inputUP := form.Find(`input`)

	for i := 0; i < 10; i++ { //input上面那一拨
		name, _ := inputUP.Eq(i).Attr(`name`)
		value, _ := inputUP.Eq(i).Attr(`value`)
		if name == "issubmit" {
			value = "1"
		}
		repdata.InputUP += name + "=" + value + "&"
	}

	form.Find(`tr`).Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			return
		}
		options := make(map[string]interface{})
		reg, _ := regexp.Compile(`\s`)
		str := reg.ReplaceAllString(selection.Find(`td`).Eq(0).Text(), "")
		if str == "教师回复：" {
			return
		}
		options["name"] = str
		options["data"] = [][]string{}
		selection.Find(`input`).Each(func(i int, selection *goquery.Selection) {
			if i%2 == 0 {
				name, _ := selection.Attr(`name`)
				value, _ := selection.Attr(`value`)
				options["data"] = append(options["data"].([][]string), []string{name, value})
			}
		})
		repdata.Options = append(repdata.Options, options)
	})
	repdata.Options = repdata.Options[:len(repdata.Options)-1]
	reply.Res = repdata
	return nil
}

func (u *User) EvaluationPost(ctx context.Context, args *Args, reply *Reply) error {
	url := args.Ip + `:8080/jsxsd/xspj/xspj_save.do`
	fmt.Println(args.Data)
	body, _, err := PostWebPage(&url, args.Session, args.Data, false)
	fmt.Println(string(body))
	reply.Res = "提交成功"
	if err != nil {
		return nil
	}
	return nil
}

type EvaRep2 struct {
	Name    string
	Teacher string
	Score   string //总评分
	Url     string
}

type RepForm struct {
	InputUP string
	Options []map[string]interface{}
}

type EvaRep1 struct {
	Name      string
	StartTime string
	EndTime   string
	Urls      []string
}
