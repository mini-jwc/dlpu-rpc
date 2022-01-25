package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	urlPak "net/url"
	"regexp"
	"strings"
)

// KeepSession
func (u *User) KeepSession(ctx context.Context, args *Args, reply *Reply) error {
	url := args.Ip + `:8080/jsxsd/framework/blankPage.jsp`
	res, _, err := GetWebPage(&url, args.Session, false)
	if err != nil {
		reply.Err = err
		return nil
	}
	fmt.Println(len(res))

	if len(res) == 6 {
		reply.Res = `更新成功`
		return nil
	}
	reply.Err = errors.New(`更新失败，session过期`)
	return nil
}

// CourseTimetable 课程表
func (u *User) CourseTimetable(ctx context.Context, args *Args, reply *Reply) error { //课程表
	url := args.Ip + `:8080/jsxsd/xskb/xskb_list.do`
	_, doc, err := PostWebPage(&url, args.Session, `xnxq01id=`+args.Data, true)
	if err != nil {
		reply.Err = err
		return nil
	}
	var ctts []CTT

	doc.Find(".kbcontent").Each(func(i int, s *goquery.Selection) {
		data2, _ := s.Html()
		if len(data2) < 5 {
			return
		}
		minusIndex := regexp.MustCompile(`---------------------`).FindAllStringIndex(data2, -1)
		teachers := s.Find(`font[title='教师']`)
		weeks := s.Find(`font[title='周次(节次)']`)
		classes := s.Find(`font[title='教室']`)
		if len(teachers.Text()) == 0 {
			teachers = s.Find(`font[title='老师']`)
		}
		var cttDetails CTTDetails

		//fmt.Println(teachers.Text(), classes.Text(), weeks.Text(), minusIndex)
		for i2, _ := range minusIndex {
			minusIndex[i2][1] += 5 //br问题
		}
		minusIndex = append([][]int{{0, 0}}, minusIndex...)
		for k := 0; k < len(minusIndex); k++ {
			b := strings.Index(data2[minusIndex[k][1]:], "<br/>")
			week := weeks.Eq(k).Text()
			weeksIndex := strings.Index(week, `[`)
			if weeksIndex != -1 {
				week = week[:weeksIndex]
			}

			cttDetails = append(cttDetails, CTTDetail{
				Name:    data2[minusIndex[k][1] : minusIndex[k][1]+b], //老师
				Teacher: teachers.Eq(k).Text(),
				Week:    week,
				Room:    classes.Eq(k).Text(),
			})
		}
		//fmt.Println(minusIndex)
		ctts = append(ctts, CTT{
			CTTDetails: cttDetails,
			Id:         i,
		})

	})
	if len(ctts) == 0 {
		reply.Res = []CTT{
			{
				CTTDetails: CTTDetails{},
				Id:         0,
			},
		}
	}
	reply.Res = ctts
	return nil
}

// ExamDate 考试日期
func (u *User) ExamDate(ctx context.Context, args *Args, reply *Reply) error {
	url := args.Ip + `:8080/jsxsd/xsks/xsksap_list`

	_, doc, err := PostWebPage(&url, args.Session, "xqlbmc=&xnxqid="+args.Data+"&xqlb=", true) //选择学期
	if err != nil {
		reply.Err = err
		return nil
	}
	var rep []ExamTime
	doc.Find(`tbody`).Eq(1).Find(`tr`).Each(func(i int, selection *goquery.Selection) {

		if i != 0 {
			selection = selection.Find(`td`)
			rep = append(rep, ExamTime{
				ID:   selection.Eq(1).Text(),
				Name: selection.Eq(3).Text(),
				Time: selection.Eq(4).Text(),
				Room: selection.Eq(5).Text(),
			})

		}
	})
	reply.Res = rep

	return nil
}

// ExamScore 考试分数
func (u *User) ExamScore(ctx context.Context, args *Args, reply *Reply) error { //考试分数
	fmt.Println(`考试分数`, len(args.Data), args.Data)
	if len(args.Data) != 11 && len(args.Data) != 0 { //分数详情
		query, err := urlPak.QueryUnescape(args.Data)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		url := args.Ip + `:8080/jsxsd/kscj/pscj_list.do?` + query
		_, doc, err := GetWebPage(&url, args.Session, true)
		if err != nil {
			reply.Err = err
			return nil
		}
		sel := doc.Find(`td`)
		if sel.Eq(7).Text() == `？？？` {
			return nil
		}

		reply.Res = ExamScoreDetail{
			Ordinary:      sel.Eq(1).Text(),
			OrdPercent:    sel.Eq(2).Text(),
			Middle:        sel.Eq(3).Text(),
			MiddlePercent: sel.Eq(4).Text(),
			Final:         sel.Eq(5).Text(),
			FinalPercent:  sel.Eq(6).Text(),
			Total:         sel.Eq(7).Text(),
		}
		return nil
	}
	rep := make(map[string]interface{})
	url := args.Ip + `:8080/jsxsd/kscj/cjcx_list`
	//res, err := PostWebPage(&url, args.Session, `kksj=2019-2020-2&bcsj=&kcxz=&kcmc=&xsfs=all`)
	//fmt.Println(args.Data)
	_, doc, err := PostWebPage(&url, args.Session, `kksj=`+args.Data+`&bcsj=&kcxz=&kcmc=&xsfs=all`, true)
	if err != nil {
		reply.Err = err
		return nil
	}

	text := doc.Find(`#Form1`).Text()
	idx := strings.Index(text, `所修课程平均学分绩点`) //找绩点
	rep["GPA"] = text[idx+33 : idx+37]
	var scores []ExamScore
	doc.Find(`tr`).Each(func(i int, selection *goquery.Selection) {
		if i > 1 {
			sec := selection.Find(`*`)

			detail, _ := sec.Eq(5).Attr(`href`)

			scores = append(scores, ExamScore{
				Name:       sec.Eq(3).Text(),
				GPA:        sec.Eq(6).Text(),
				Detail:     detail[43 : len(detail)-10],
				Grade:      sec.Eq(4).Text(),
				Credit:     sec.Eq(8).Text(),
				Period:     sec.Eq(9).Text(),
				Property:   sec.Eq(11).Text(),
				Properties: sec.Eq(12).Text(),
				BC:         sec.Eq(13).Text(),
			})

		}
	})
	rep["Scores"] = scores
	reply.Res = rep
	return nil
}

// CultivateScheme 培养方案
func (u *User) CultivateScheme(ctx context.Context, args *Args, reply *Reply) error { //培养方案
	url := args.Ip + `:8080/jsxsd/pyfa/pyfazd_query`
	_, doc, err := GetWebPage(&url, args.Session, true)
	if err != nil {
		reply.Err = err
		return nil
	}
	var rep []CultivateScheme
	doc.Find(`tr`).Each(func(i int, selection *goquery.Selection) {
		if i > 1 {
			sec := selection.Find(`td`)
			rep = append(rep, CultivateScheme{
				Semester:   sec.Eq(1).Text(),
				Name:       sec.Eq(3).Text(),
				Credit:     sec.Eq(5).Text(),
				ExamMode:   sec.Eq(6).Text(),
				Period:     sec.Eq(4).Text(),
				WeekPeriod: sec.Eq(8).Text(),
				College:    sec.Eq(7).Text(),
			})
		}
	})
	reply.Res = rep
	return nil
}

//GetWebPage 封装get请求
func GetWebPage(url *string, session string, doc bool) ([]byte, *goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		return nil, nil, err
		// handle error
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+session)
	//fmt.Println(*url, session)

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	err = checkPage(res)
	if doc {
		d, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			fmt.Println(err)
			return nil, nil, err
		}
		return nil, d, nil

	} else {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
		return b, nil, nil

	}
}

//PostWebPage 封装post请求 doc 是否返回doc
func PostWebPage(url *string, session string, data string, doc bool) ([]byte, *goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", *url, strings.NewReader(data))
	if req == nil {
		return nil, nil, err
		// handle error
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Cookie", "JSESSIONID="+session)
	//fmt.Println(*url, session)
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	err = checkPage(res)
	if err != nil {
		return nil, nil, err
	}

	if doc {
		d, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			fmt.Println(err)
			return nil, nil, err
		}
		return nil, d, nil

	} else {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
		return b, nil, nil

	}

}
func checkPage(res *http.Response) error {
	if res.Header.Get(`Content-Type`) == "text/html;charset=GBK" {
		fmt.Println(`session过期`)
		return errors.New(`session过期`)
	}
	return nil
}
