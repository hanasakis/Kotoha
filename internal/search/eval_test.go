package search

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
	"github.com/hanasakis/kotoha/pkg/embedding"
	"github.com/hanasakis/kotoha/pkg/milvus"
)

type searchEvalQuery struct {
	ID          string `json:"id"`
	QueryZH     string `json:"query_zh"`
	QueryEN     string `json:"query_en"`
	ExpectedIDs []int  `json:"expected_ids"`
	Scene       string `json:"scene"`
	Taste       string `json:"taste"`
	MinRelevant int    `json:"min_relevant"`
	Notes       string `json:"notes"`
}

type searchEvalMetrics struct {
	RecallAtK   []int   `json:"recall_at_k"`
	MRR         bool    `json:"mrr"`
	NDCG        bool    `json:"ndcg"`
	DenseWeight float64 `json:"dense_weight"`
	BM25Weight  float64 `json:"bm25_weight"`
}

type searchEvalFile struct {
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Queries     []searchEvalQuery `json:"queries"`
	Metrics     searchEvalMetrics `json:"metrics"`
}

func loadSearchEvals(t *testing.T) []searchEvalQuery {
	t.Helper()

	candidates := []string{
		"evaluation/search_queries.json",
		"../../evaluation/search_queries.json",
		filepath.Join("..", "..", "evaluation", "search_queries.json"),
	}

	var path string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			path = c
			break
		}
	}
	if path == "" {
		t.Skip("search_queries.json not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read eval file: %v", err)
	}

	var ef searchEvalFile
	if err := json.Unmarshal(data, &ef); err != nil {
		t.Fatalf("failed to parse eval file: %v", err)
	}
	return ef.Queries
}

func loadSearchEvalFile(t *testing.T) *searchEvalFile {
	t.Helper()

	candidates := []string{
		"evaluation/search_queries.json",
		"../../evaluation/search_queries.json",
		filepath.Join("..", "..", "evaluation", "search_queries.json"),
	}

	var path string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			path = c
			break
		}
	}
	if path == "" {
		t.Skip("search_queries.json not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read eval file: %v", err)
	}

	var ef searchEvalFile
	if err := json.Unmarshal(data, &ef); err != nil {
		t.Fatalf("failed to parse eval file: %v", err)
	}
	return &ef
}

// computeMRR returns Mean Reciprocal Rank: average of 1/rank of the first relevant result.
// Returns 0 if no relevant results are found.
func computeMRR(results []catalog.Product, expectedIDs []int) float64 {
	relevant := make(map[int]bool, len(expectedIDs))
	for _, id := range expectedIDs {
		relevant[id] = true
	}
	for i, r := range results {
		if relevant[int(r.ID)] {
			return 1.0 / float64(i+1)
		}
	}
	return 0.0
}

// computeNDCG returns Normalized Discounted Cumulative Gain at position k.
// Uses binary relevance (1 if in expected set, 0 otherwise).
func computeNDCG(results []catalog.Product, expectedIDs []int, k int) float64 {
	if k == 0 {
		return 0.0
	}
	relevant := make(map[int]bool, len(expectedIDs))
	for _, id := range expectedIDs {
		relevant[id] = true
	}

	// DCG@k
	var dcg float64
	n := k
	if n > len(results) {
		n = len(results)
	}
	for i := 0; i < n; i++ {
		rel := 0.0
		if relevant[int(results[i].ID)] {
			rel = 1.0
		}
		dcg += rel / math.Log2(float64(i+2))
	}

	// IDCG@k
	var idcg float64
	idealN := len(expectedIDs)
	if idealN > k {
		idealN = k
	}
	for i := 0; i < idealN; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}
	if idcg == 0 {
		return 0.0
	}
	return dcg / idcg
}

func TestEvalSearch_CatalogCoverage(t *testing.T) {
	queries := loadSearchEvals(t)

	pg := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(pg); err != nil {
		t.Fatalf("auto-migrate failed: %v", err)
	}
	defer testutil.CleanTestDB(t, pg)

	catalogRepo := catalog.NewRepository(pg)
	catalogSvc := catalog.NewService(catalogRepo, "http://localhost:9000/kotoha-images/products/")
	if err := catalogSvc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	// Collect all unique expected product IDs across queries
	allIDs := make(map[int]bool)
	for _, q := range queries {
		for _, id := range q.ExpectedIDs {
			allIDs[id] = true
		}
	}

	// Verify each expected ID exists in the catalog
	for id := range allIDs {
		_, err := catalogRepo.GetProduct(uint(id))
		if err != nil {
			t.Errorf("expected product %d not found: %v (referenced by eval queries)", id, err)
		}
	}
	t.Logf("All %d expected product IDs verified in catalog", len(allIDs))
}

func TestEvalSearch_Queries(t *testing.T) {
	evalFile := loadSearchEvalFile(t)
	queries := evalFile.Queries

	pg := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(pg); err != nil {
		t.Fatalf("auto-migrate failed: %v", err)
	}
	defer testutil.CleanTestDB(t, pg)

	catalogRepo := catalog.NewRepository(pg)
	catalogSvc := catalog.NewService(catalogRepo, "http://localhost:9000/kotoha-images/products/")
	if err := catalogSvc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("cannot load config: %v", err)
	}

	milvusCli := milvus.New(cfg.MilvusAddr(), cfg.Milvus.DB, cfg.Milvus.User, cfg.Milvus.Password)
	embeddingCli := embedding.New(cfg.Ollama.Host, cfg.Embedding.Model)
	svc := NewService(milvusCli, embeddingCli, catalogRepo, cfg.Milvus.DenseWeight, cfg.Milvus.BM25Weight, cfg.Milvus.RecallTopK, cfg.Embedding.Dim)

	// Index all products (skip if Milvus unavailable)
	if err := svc.IndexAll(); err != nil {
		t.Skipf("skipping search eval: cannot index (Milvus/Ollama likely down): %v", err)
	}

	passed := 0
	failed := 0
	var (
		totalMRR    float64
		totalNDCG5  float64
		totalNDCG10 float64
		totalRecall float64
		queryCount  int
	)

	for _, q := range queries {
		t.Run(q.ID, func(t *testing.T) {
			results, err := svc.Search(q.QueryZH, 10)
			if err != nil {
				t.Errorf("[%s] %s: search error: %v", q.ID, q.QueryZH, err)
				failed++
				return
			}

			resultIDs := make(map[uint]bool)
			for _, r := range results {
				resultIDs[r.ID] = true
			}

			matched := 0
			for _, expectedID := range q.ExpectedIDs {
				if resultIDs[uint(expectedID)] {
					matched++
				}
			}

			recall := float64(matched) / float64(len(q.ExpectedIDs))
			mrr := computeMRR(results, q.ExpectedIDs)
			ndcg5 := computeNDCG(results, q.ExpectedIDs, 5)
			ndcg10 := computeNDCG(results, q.ExpectedIDs, 10)

			totalMRR += mrr
			totalNDCG5 += ndcg5
			totalNDCG10 += ndcg10
			totalRecall += recall
			queryCount++

			if matched < q.MinRelevant {
				t.Errorf("[%s] %s: recall=%.2f mrr=%.4f ndcg@5=%.4f ndcg@10=%.4f (%d/%d matched), min=%d",
					q.ID, q.QueryZH, recall, mrr, ndcg5, ndcg10, matched, len(q.ExpectedIDs), q.MinRelevant)
				failed++
			} else {
				t.Logf("[%s] %s: PASS recall=%.2f mrr=%.4f ndcg@5=%.4f ndcg@10=%.4f (%d/%d matched)",
					q.ID, q.QueryZH, recall, mrr, ndcg5, ndcg10, matched, len(q.ExpectedIDs))
				passed++
			}
		})
	}

	t.Logf("=== Search Eval Summary ===")
	t.Logf("Total: %d | Passed: %d | Failed: %d", len(queries), passed, failed)
	if queryCount > 0 {
		avgMRR := totalMRR / float64(queryCount)
		avgNDCG5 := totalNDCG5 / float64(queryCount)
		avgNDCG10 := totalNDCG10 / float64(queryCount)
		avgRecall := totalRecall / float64(queryCount)

		t.Logf("=== Ranking Metrics ===")
		t.Logf("MRR: %.4f | NDCG@5: %.4f | NDCG@10: %.4f | Avg Recall: %.2f",
			avgMRR, avgNDCG5, avgNDCG10, avgRecall)
	}

	accuracy := 0.0
	if passed+failed > 0 {
		accuracy = float64(passed) / float64(passed+failed) * 100
		t.Logf("Accuracy: %d/%d (%.1f%%)", passed, passed+failed, accuracy)
	}

	// Emit machine-parseable result line for CI pipeline
	if queryCount > 0 {
		avgMRR := totalMRR / float64(queryCount)
		avgNDCG5 := totalNDCG5 / float64(queryCount)
		avgNDCG10 := totalNDCG10 / float64(queryCount)
		avgRecall := totalRecall / float64(queryCount)

		metricsJSON, _ := json.Marshal(map[string]interface{}{
			"mrr":        avgMRR,
			"ndcg_at_5":  avgNDCG5,
			"ndcg_at_10": avgNDCG10,
			"avg_recall": avgRecall,
			"total":      len(queries),
			"passed":     passed,
			"failed":     failed,
			"accuracy":   accuracy,
		})
		t.Logf("EVAL_RESULT: %s", string(metricsJSON))
	}
}
