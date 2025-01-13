package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"golang.org/x/net/proxy"
)

type Config struct {
	JiraURL     string `json:"jira_url"`
	JiraToken   string `json:"jira_token"`
	Socks5Proxy string `json:"socks5_proxy"`
	WebhookURL  string `json:"webhook_url"`
}

type WebhookPayload struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func loadConfig(filePath string) (Config, error) {
	var config Config
	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func sendDingTalkMessage(webhookURL, content string) error {
	payload := WebhookPayload{
		MsgType: "text",
	}
	payload.Text.Content = content

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error sending webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send webhook: %d", resp.StatusCode)
	}

	return nil
}

// BearerAuthTransport is a custom transport to add Bearer Token authentication
type BearerAuthTransport struct {
	Token     string
	Transport http.RoundTripper
}

func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	return t.Transport.RoundTrip(req)
}

func sendMessages(client *jira.Client, config Config, currentUser *jira.User) {
	// 获取当天日期
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	threeDaysLater := time.Now().AddDate(0, 0, 3).Format("2006-01-02")

	// 搜索当天到期的任务
	jqlToday := fmt.Sprintf(`duedate = "%s" AND assignee = "%s"`, today, currentUser.Name)
	issuesToday, _, err := client.Issue.Search(jqlToday, nil)
	if err != nil {
		fmt.Printf("Error searching issues for today: %v\n", err)
		return
	}

	// 搜索未来三天到期的任务
	jqlFuture := fmt.Sprintf(`duedate >= "%s" AND duedate <= "%s" AND assignee = "%s"`, tomorrow, threeDaysLater, currentUser.Name)
	issuesFuture, _, err := client.Issue.Search(jqlFuture, nil)
	if err != nil {
		fmt.Printf("Error searching issues for the next three days: %v\n", err)
		return
	}

	// 构建消息内容
	var content string
	content += fmt.Sprintf("今天 %s 到期任务\n", today)
	for _, issue := range issuesToday {
		content += fmt.Sprintf("任务: %s - %s\n", issue.Key, issue.Fields.Summary)
	}

	content += "\n未来三天到期任务\n"
	for _, issue := range issuesFuture {
		content += fmt.Sprintf("任务: %s - %s\n", issue.Key, issue.Fields.Summary)
	}

	// 发送消息到钉钉
	if err := sendDingTalkMessage(config.WebhookURL, content); err != nil {
		fmt.Printf("Error sending message to DingTalk: %v\n", err)
		return
	}

	fmt.Println("Message sent to DingTalk successfully")
}

func main() {
	// 加载配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// 配置SOCKS5代理
	dialer, err := proxy.SOCKS5("tcp", config.Socks5Proxy, nil, proxy.Direct)
	if err != nil {
		fmt.Printf("Error creating SOCKS5 dialer: %v\n", err)
		return
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	// 使用 Bearer Token 初始化 JIRA 客户端
	bearerTransport := &BearerAuthTransport{
		Token:     config.JiraToken,
		Transport: transport,
	}

	client, err := jira.NewClient(&http.Client{Transport: bearerTransport}, config.JiraURL)
	if err != nil {
		fmt.Printf("Error creating JIRA client: %v\n", err)
		return
	}

	// 获取当前用户名
	currentUser, _, err := client.User.GetSelf()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return
	}
	sendMessages(client, config, currentUser)
	for {
		now := time.Now()
		next8AM := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		next6PM := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())

		if now.After(next8AM) {
			next8AM = next8AM.AddDate(0, 0, 1)
		}
		if now.After(next6PM) {
			next6PM = next6PM.AddDate(0, 0, 1)
		}

		if next8AM.Before(next6PM) {
			time.Sleep(next8AM.Sub(now))
			sendMessages(client, config, currentUser)
		} else {
			time.Sleep(next6PM.Sub(now))
			sendMessages(client, config, currentUser)
		}
	}
}
