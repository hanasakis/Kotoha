package milvus

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	dbName  string
	authHdr string
	httpCli *http.Client
}

func New(addr, dbName, user, password string) *Client {
	authHdr := ""
	if user != "" {
		authHdr = "Bearer " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
	}
	return &Client{
		baseURL: "http://" + addr + "/v2/vectordb",
		dbName:  dbName,
		authHdr: authHdr,
		httpCli: &http.Client{Timeout: 60 * time.Second},
	}
}

type collectionSchema struct {
	CollectionName string        `json:"collectionName"`
	DBName         string        `json:"dbName"`
	Dimension      int           `json:"dimension"`
	MetricType     string        `json:"metricType"`
	PrimaryField   string        `json:"primaryField"`
	VectorField    string        `json:"vectorField"`
	EnableDynamic  bool          `json:"enableDynamic"`
}

type createReq struct {
	CollectionName string `json:"collectionName"`
	DBName         string `json:"dbName"`
	Dimension      int    `json:"dimension"`
	MetricType     string `json:"metricType"`
	PrimaryField   string `json:"primaryField"`
	VectorField    string `json:"vectorField"`
	EnableDynamic  bool   `json:"enableDynamic"`
}

func (c *Client) doPost(path string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if c.authHdr != "" {
		req.Header.Set("Authorization", c.authHdr)
	}
	return c.httpCli.Do(req)
}

func (c *Client) doGet(path string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", c.baseURL+path, nil)
	if c.authHdr != "" {
		req.Header.Set("Authorization", c.authHdr)
	}
	return c.httpCli.Do(req)
}

func (c *Client) CreateCollection(name string, dim int) error {
	req := createReq{
		CollectionName: name,
		DBName:         c.dbName,
		Dimension:      dim,
		MetricType:     "IP",
		PrimaryField:   "id",
		VectorField:    "vector",
		EnableDynamic:  true,
	}
	body, _ := json.Marshal(req)

	resp, err := c.doPost("/collections/create", body)
	if err != nil {
		return fmt.Errorf("milvus.create_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct{ Message string }
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message != "" {
			return fmt.Errorf("milvus.create_error: %s", errResp.Message)
		}
		return fmt.Errorf("milvus.create_error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) HasCollection(name string) (bool, error) {
	resp, err := c.doGet("/collections/describe?collectionName=" + name + "&dbName=" + c.dbName)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

type insertReq struct {
	CollectionName string              `json:"collectionName"`
	DBName         string              `json:"dbName"`
	Data           []map[string]interface{} `json:"data"`
}

func (c *Client) Insert(collectionName string, rows []map[string]interface{}) error {
	req := insertReq{
		CollectionName: collectionName,
		DBName:         c.dbName,
		Data:           rows,
	}
	body, _ := json.Marshal(req)

	resp, err := c.doPost("/entities/insert", body)
	if err != nil {
		return fmt.Errorf("milvus.insert_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("milvus.insert_error: status %d", resp.StatusCode)
	}
	return nil
}

type queryReq struct {
	CollectionName string   `json:"collectionName"`
	DBName         string   `json:"dbName"`
	Filter         string   `json:"filter"`
	Limit          int      `json:"limit"`
	OutputFields   []string `json:"outputFields"`
}

func (c *Client) QueryByKeyword(collectionName string, keywords []string, limit int) ([]int64, error) {
	escaped := make([]string, len(keywords))
	for i, kw := range keywords {
		kw = strings.ReplaceAll(kw, "'", "''")
		kw = strings.ReplaceAll(kw, `"`, `\"`)
		escaped[i] = kw
	}
	var conditions []string
	for _, kw := range escaped {
		conditions = append(conditions, fmt.Sprintf(`text like "%%%s%%"`, kw))
	}
	filter := strings.Join(conditions, " or ")

	req := queryReq{
		CollectionName: collectionName,
		DBName:         c.dbName,
		Filter:         filter,
		Limit:          limit,
		OutputFields:   []string{"id"},
	}
	body, _ := json.Marshal(req)

	resp, err := c.doPost("/entities/query", body)
	if err != nil {
		return nil, fmt.Errorf("milvus.query_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("milvus.query_error: status %d", resp.StatusCode)
	}

	var rawResp struct {
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("milvus.decode_error: %w", err)
	}

	ids := make([]int64, len(rawResp.Data))
	for i, d := range rawResp.Data {
		ids[i] = d.ID
	}
	return ids, nil
}

type searchReq struct {
	CollectionName string        `json:"collectionName"`
	DBName         string        `json:"dbName"`
	Data           [][]float64   `json:"data"`
	AnnsField      string        `json:"annsField"`
	Limit          int           `json:"limit"`
	OutputFields   []string      `json:"outputFields"`
}

type searchResp struct {
	Data []struct {
		ID       int64              `json:"id"`
		Distance float64            `json:"distance"`
		Fields   map[string]interface{} `json:"-"`
	} `json:"data"`
}

func (c *Client) Search(collectionName string, vector []float64, limit int, outputFields []string) ([]SearchResult, error) {
	req := searchReq{
		CollectionName: collectionName,
		DBName:         c.dbName,
		Data:           [][]float64{vector},
		AnnsField:      "vector",
		Limit:          limit,
		OutputFields:   outputFields,
	}
	body, _ := json.Marshal(req)

	resp, err := c.doPost("/entities/search", body)
	if err != nil {
		return nil, fmt.Errorf("milvus.search_error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("milvus.search_error: status %d", resp.StatusCode)
	}

	var rawResp struct {
		Data []struct {
			ID       int64                  `json:"id"`
			Distance float64                `json:"distance"`
			Entity   map[string]interface{} `json:"entity"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("milvus.decode_error: %w", err)
	}

	results := make([]SearchResult, len(rawResp.Data))
	for i, d := range rawResp.Data {
		results[i] = SearchResult{
			ID:       d.ID,
			Distance: d.Distance,
			Fields:   d.Entity,
		}
	}
	return results, nil
}

type SearchResult struct {
	ID       int64
	Distance float64
	Fields   map[string]interface{}
}
