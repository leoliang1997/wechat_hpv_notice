package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type HospitalResp struct {
	List   []HospitalInfo `json:"list"`
	Status int            `json:"status"`
	Msg    string         `json:"msg"`
}

type HospitalInfo struct {
	ID           int           `json:"id"`
	Cname        string        `json:"cname"`
	Addr         string        `json:"addr"`
	SmallPic     string        `json:"SmallPic"`
	BigPic       interface{}   `json:"BigPic"`
	Lat          float64       `json:"lat"`
	Lng          float64       `json:"lng"`
	Tel          string        `json:"tel"`
	Addr2        string        `json:"addr2"`
	Province     int           `json:"province"`
	City         int           `json:"city"`
	County       int           `json:"county"`
	Sort         int           `json:"sort"`
	DistanceShow int           `json:"DistanceShow"`
	PayMent      string        `json:"PayMent"`
	IdcardLimit  bool          `json:"IdcardLimit"`
	Notice       string        `json:"notice"`
	Distance     float64       `json:"distance"`
	Tags         []interface{} `json:"tags"`
}

type HospitalDetail struct {
	Tel         string    `json:"tel"`
	Addr        string    `json:"addr"`
	Cname       string    `json:"cname"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Distance    int       `json:"distance"`
	Payment     Payment   `json:"payment"`
	BigPic      string    `json:"BigPic"`
	IdcardLimit bool      `json:"IdcardLimit"`
	Notice      string    `json:"notice"`
	Status      int       `json:"status"`
	List        []HPVInfo `json:"list"`
}

type Payment struct {
	Alipay    string `json:"alipay"`
	WechatPay string `json:"WechatPay"`
	UnionPay  string `json:"UnionPay"`
	Cashier   string `json:"cashier"`
}

type NumbersVaccine struct {
	Cname string `json:"cname"`
	Value int    `json:"value"`
}

type HPVInfo struct {
	ID              int              `json:"id"`
	Text            string           `json:"text"`
	Price           string           `json:"price"`
	Descript        string           `json:"descript"`
	Warn            string           `json:"warn"`
	Tags            []string         `json:"tags"`
	QuestionnaireID int              `json:"questionnaireId"`
	Remarks         string           `json:"remarks"`
	NumbersVaccine  []NumbersVaccine `json:"NumbersVaccine"`
	Date            string           `json:"date"`
	BtnLable        string           `json:"BtnLable"`
	Enable          bool             `json:"enable"`
}

type Notice struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

type Text struct {
	Content string `json:"content"`
}

func FetchHospitalList() (*HospitalResp, error) {
	url := "https://cloud.cn2030.com/sc/wx/HandlerSubscribe.ashx?act=CustomerList&city=%5B%22%E5%B9%BF%E8%A5%BF%E5%A3%AE%E6%97%8F%E8%87%AA%E6%B2%BB%E5%8C%BA%22%2C%22%E5%8D%97%E5%AE%81%E5%B8%82%22%2C%22%22%5D&lat=22.547216796875&lng=113.94323920355903&id=0&cityCode=450100&product=1"
	body, err := fetchResBody(url)

	var res HospitalResp
	err = json.Unmarshal(body, &res)
	return &res, err
}

func FetchHPVInfo(h HospitalInfo) (*HospitalDetail, error) {
	url := "https://cloud.cn2030.com/sc/wx/HandlerSubscribe.ashx?act=CustomerProduct&lat=22.547216796875&lng=113.94323920355903&id=" + strconv.Itoa(h.ID)
	body, err := fetchResBody(url)

	var res HospitalDetail
	err = json.Unmarshal(body, &res)
	return &res, err
}

func fetchResBody(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", "https://servicewechat.com/wx2c7f0f3c30d99445/91/page-frame.html")
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.16(0x18001032) NetType/WIFI Language/zh_CN")
	req.Header.Add("Host", "cloud.cn2030.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func SendWechatMsg(hospitalName, tel, addr, title, btnLabel string) error {
	url := "https://oapi.dingtalk.com/robot/send?access_token=a58250aeb6f685b253f81f04b8f60e45ac02e022a7a28529a2b200cabf76053e"
	client := &http.Client{Timeout: 5 * time.Second}
	content := "医院名: " + hospitalName + "\n" +
		"电话: " + tel + "\n" +
		"地址: " + addr + "\n" +
		"标题: " + title + "\n" +
		"状态: " + btnLabel
	notice := Notice{MsgType: "text", Text: Text{Content: content}}
	s, _ := json.Marshal(notice)

	req, err := http.NewRequest("POST", url, bytes.NewReader(s))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res, _ := io.ReadAll(resp.Body)
	fmt.Println(string(res))
	return nil
}

var hospitalInfo atomic.Value

func main() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("程序panic%#v", r)
				os.Exit(-1)
			}
		}()

		for {
			h, err := FetchHospitalList()
			if err != nil {
				fmt.Println("无法获取医院列表，系统退出")
				os.Exit(-1)
			}
			hospitalInfo.Store(h)
			time.Sleep(5 * time.Second) //每五秒刷新一次
		}
	}()

	for {
		h, ok := hospitalInfo.Load().(*HospitalResp)
		if !ok {
			time.Sleep(5 * time.Second)
			continue
		}
		for _, list := range h.List {
			time.Sleep(1 * time.Second)
			detail, err := FetchHPVInfo(list)
			if err != nil {
				fmt.Println("拉取医院详情失败", err)
				continue
			}

			fmt.Println("医院名: " + detail.Cname + "  电话: " + detail.Tel)
			for _, info := range detail.List {
				if strings.Contains(info.Text, "九价") {
					fmt.Println(info.Text + ": " + info.BtnLable)
					if info.Enable {
						SendWechatMsg(detail.Cname, detail.Tel, detail.Addr, info.Text, info.BtnLable)
					}
				}
			}
			fmt.Println()
		}
	}
}
