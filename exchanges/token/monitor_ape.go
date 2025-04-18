package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultEndpoint = "https://api.ape.pro/api/v1/gems"
	defaultTimeout  = 30 * time.Second
)

// Client Ape Pro API客户端
type Client struct {
	endpoint string
	client   *http.Client
	headers  map[string]string
}

// NewClient 创建新的API客户端
func NewClient() *Client {
	return &Client{
		endpoint: defaultEndpoint,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		headers: map[string]string{
			"accept":             "*/*",
			"accept-language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
			"content-type":       "application/json",
			"cookie":             "cf_clearance=9aYSJi1nYB46ZLW37KvvVevlkLmDvlzce5XGwPZE9VM-1736822574-1.2.1.1-3h4a52VprdVeakKe3tZQY6sHimRoxD20JYaCtqSYOmsqiq_q542pOnPFe2RHuFB1CZoCYhV7dhnnO8E5jeDAoEHczmPtUtRuoYKaJ31w490UxscOHm.DeFHC0CJDt9s5t58ul4AyfhwRkSa4Lgm0GbkrocwgeVP4xf1kvpWt2_KYb5VvZIrZnr4fMMPSk6eQKKvfpkfOcYuC219XtH87cLRtS9Y5CeEvXPID5Fw0JsQ7doGC51zaayHpO1fbmH50caON0Fo9BqAMhALQQc0yW56LoF.w_wM781RnmtOrqrdSlVB.MtyPzvsmM2V4CqMC1hjLlcrjQoUoKiFiQF0aZw",
			"dnt":                "1",
			"origin":             "https://ape.pro",
			"priority":           "u=1, i",
			"referer":            "https://ape.pro/",
			"sec-ch-ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
			"sec-ch-ua-mobile":   "?1",
			"sec-ch-ua-platform": `"Android"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-site",
			"user-agent":         "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
			"x-ape-client-id":    "529914d",
			"x-ape-client-token": "v3FGgD6Ensof6kCno2dz3fDTVqylFayiz1M75v2v1B5+D+b2+pd8TRU3+8W8xV/v91n6wi99+NfHtypfV9nK9thAjYolLDoey8w++rv/HOthiXNTsJ8Zhzc=",
		},
	}
}

// WithEndpoint 设置自定义API端点
func (c *Client) WithEndpoint(endpoint string) *Client {
	c.endpoint = endpoint
	return c
}

// WithTimeout 设置自定义超时时间
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.client.Timeout = timeout
	return c
}

// WithHeader 添加或修改请求头
func (c *Client) WithHeader(key, value string) *Client {
	c.headers[key] = value
	return c
}

// GetGemData 获取并分类代币数据
func (c *Client) GetGemData() (*APIResponse, error) {
	body := `{"new":{"notPumpfunToken":false},"aboutToGraduate":{},"graduated":{}}`
	bodyReader := bytes.NewReader([]byte(body))

	req, err := http.NewRequest("POST", c.endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("非预期状态码: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &apiResponse, nil
}

// GetAllPools 获取所有分类合并后的池列表
func (c *Client) GetAllPools() ([]Pool, error) {
	response, err := c.GetGemData()
	if err != nil {
		return nil, err
	}

	var allPools []Pool
	allPools = append(allPools, response.Graduated.Pools...)

	return allPools, nil
}
