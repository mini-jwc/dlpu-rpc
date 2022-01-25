package model

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	urlPak "net/url"
	"regexp"
	"strings"
)

func (u *User) Login(ctx context.Context, args *Args, reply *Reply) error {

	client := &http.Client{
		CheckRedirect: checkRedirect,
	}
	fmt.Println(`登录`)
	//fmt.Println(args.StuId, args.Pwd)
	v := urlPak.Values{}
	v.Add("USERNAME", args.StuId)
	v.Add("PASSWORD", args.Pwd)
	fmt.Println(v.Encode())
	req, err := http.NewRequest("POST", args.Ip+":8080/jsxsd/xk/LoginToXk", strings.NewReader(v.Encode()))
	if req == nil {
		return errors.New(`请求错误`)
		// handle error
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(resp.Status)
	if resp.Status[0:3] == "302" {
		//重定向，登录成功
		//fmt.Println(resp.Header.Get(`Set-Cookie`))
		reply.Res = map[string]string{"session": resp.Header.Get(`Set-Cookie`)[11:43]}
	} else {
		reply.Err = errors.New(`用户名或密码错误`)
	}
	return nil
}

// GetUserinfo 获取身份证等信息
func (u *User) GetUserinfo(ctx context.Context, args *Args, reply *Reply) error {
	url := args.Ip + `:8080/jsxsd/grxx/xsxx`

	_, doc, err := GetWebPage(&url, args.Session, true) //选择学期
	if err != nil {
		reply.Err = err
		return nil
	}
	sel := doc.Find(`#xjkpTable`).Find(`*`)

	if len(sel.Eq(13).Text()) > 9 && len(sel.Eq(14).Text()) > 9 && len(sel.Eq(241).Text()[2:]) > 9 {
		reply.Res = map[string]interface{}{
			"name":       sel.Eq(20).Text()[2:],
			"department": sel.Eq(13).Text()[9:],
			"major":      sel.Eq(14).Text()[9:],
			"idCard":     sel.Eq(241).Text()[2:],
			"class":      sel.Eq(16).Text()[9:],
		}
	} else {
		reply.Res = map[string]interface{}{
			"name":       "",
			"department": "",
			"major":      "",
			"idCard":     "",
			"class":      "",
		}
	}

	return nil
}

func (u *User) GetName(ctx context.Context, args *Args, reply *Reply) error {
	url := args.Ip + `:8080/jsxsd/framework/main.jsp`
	body, _, err := GetWebPage(&url, args.Session, false)
	if err != nil {
		fmt.Println(err.Error())
	}

	name := regexp.MustCompile("style=\"color: #000000;\">.*?\\(").FindString(string(body))

	if name != `` {
		name = name[24 : len(name)-1]
		reply.Res = name
	}
	return nil
}

func checkRedirect(req *http.Request, via []*http.Request) error { //组织自动重定向
	//自用，将url根据需求进行组合
	if len(via) >= 1 {
		return errors.New("登录成功")
	}
	return nil
}
