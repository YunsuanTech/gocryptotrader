package token

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// GMGNFetcher 用于获取GMGN数据的结构体
type GMGNFetcher struct {
	client    tls_client.HttpClient
	cookie    string
	userAgent string
}

// NewGMGNFetcher 创建一个新的GMGN数据获取器
func NewGMGNFetcher() (*GMGNFetcher, error) {
	// 初始化 HTTP 客户端
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	// 使用NoopLogger避免日志输出
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
	}

	return &GMGNFetcher{
		client:    client,
		cookie:    "cf_clearance=ErI2quTrZjRnEfDtS_tfjSLcX6bzooZlKHE2jXc2l3c-1746667491-1.2.1.1-5RF8qerYLnpQaaAIqCE.fpOfuPiGzlCtluZnqAeyC9XXkx4NrKYsgOKg5GLvQS0Bz4QcP3UFT.bifjGyPHitPK8A7_4ircc.qlb7A95gth2AyLTuiT7xD9ELJhLwNlGqTPjZrOSEVhfhQ5K2G2lgIcKpv6F5sYVylwCMfeeZgCMD.d7mtVP8W2xQ.E.QL49gPTTgDz6AhVD1n3kTK1Z0_qEdts2UudlYxTkrZ_KI5WTs2WTWFsNKCTYMP4tgTqIhYAopoq_b36L.72hw3WSDwU0lJUqO2CTw9JLZCLrzlnVuKdGJmLyQ9q.oHw31m74Kih3sBuFq5BsTM5sgCmcBdBIzM8okkZZ9KUhWbrfsBSbsKGakAvws5j9F9_cTFHTA",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	}, nil
}

// SetCookie 设置请求的cookie
func (f *GMGNFetcher) SetCookie(cookie string) {
	f.cookie = cookie
}

// SetUserAgent 设置请求的User-Agent
func (f *GMGNFetcher) SetUserAgent(userAgent string) {
	f.userAgent = userAgent
}

// GetLatestCompletedTokens 获取最新的已完成代币列表
func (f *GMGNFetcher) GetLatestCompletedTokens() ([]GMGNToken, error) {
	// 检查客户端是否已初始化
	if f.client == nil {
		return nil, fmt.Errorf("HTTP客户端未初始化")
	}

	// 创建 HTTP 请求，添加完整的查询参数
	req, err := http.NewRequest(http.MethodGet, "https://gmgn.ai/defi/quotation/v1/rank/sol/pump_ranks/1h?new_creation=%7B%22filters%22:[%22not_wash_trading%22],%22limit%22:0%7D&pump=%7B%22filters%22:[%22not_wash_trading%22],%22limit%22:0%7D&completed=%7B%22filters%22:[%22not_wash_trading%22],%22limit%22:5%7D", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header = http.Header{
		"cookie":     {f.cookie},
		"user-agent": {f.userAgent},
	}

	// 发送请求
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("收到非200状态码: %d", resp.StatusCode)
	}

	// 读取响应数据
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析 JSON 数据
	var response GMGNResponse
	if err := json.Unmarshal(readBytes, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return response.Data.Completeds, nil
}

// GetTokenKlineData 获取指定代币的K线数据
func (f *GMGNFetcher) GetTokenKlineData(tokenAddress string) ([]KlineData, error) {
	// 检查客户端是否已初始化
	if f.client == nil {
		return nil, fmt.Errorf("HTTP客户端未初始化")
	}

	// 动态生成时间参数
	now := time.Now().Unix()
	interval := 60 * 60      // 1小时间隔（单位：秒）
	lookback := interval * 4 // 需要获取4个时间间隔

	// 构造请求URL
	url := fmt.Sprintf(
		"https://www.gmgn.cc/defi/quotation/v1/tokens/kline/sol/%s?resolution=1h&from=%d&to=%d",
		tokenAddress,
		now-int64(lookback),
		now,
	)

	// 创建 HTTP 请求
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header = http.Header{
		"cookie":     {f.cookie},
		"user-agent": {f.userAgent},
	}

	// 发送请求
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("收到非200状态码: %d", resp.StatusCode)
	}

	// 读取响应数据
	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 解析 JSON 数据
	var response KlineResponse
	if err := json.Unmarshal(readBytes, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return response.Data, nil
}
