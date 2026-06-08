package search

import (
	"fmt"
	"sort"
	"strings"

	klog "github.com/hanasakis/kotoha/pkg/log"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/pkg/embedding"
	"github.com/hanasakis/kotoha/pkg/milvus"
)

const collectionName = "kotoha_products"

type Service struct {
	milvusCli    *milvus.Client
	embeddingCli *embedding.Client
	catalogRepo  *catalog.Repository
	denseWeight  float64
	bm25Weight   float64
	recallTopK   int
	embeddingDim int
}

func NewService(milvusCli *milvus.Client, embeddingCli *embedding.Client, catalogRepo *catalog.Repository, denseWeight, bm25Weight float64, recallTopK, embeddingDim int) *Service {
	return &Service{
		milvusCli:    milvusCli,
		embeddingCli: embeddingCli,
		catalogRepo:  catalogRepo,
		denseWeight:  denseWeight,
		bm25Weight:   bm25Weight,
		recallTopK:   recallTopK,
		embeddingDim: embeddingDim,
	}
}

func (s *Service) InitCollection(dim int) error {
	has, err := s.milvusCli.HasCollection(collectionName)
	if err != nil {
		return fmt.Errorf("search.has_collection_error: %w", err)
	}
	if has {
		return nil
	}
	klog.Infof("[search] creating Milvus collection: %s (dim=%d)", collectionName, dim)
	return s.milvusCli.CreateCollection(collectionName, dim)
}

type ProductDoc struct {
	ProductID uint
	Name      string
	NameEn    string
	Tags      string
	Scenes    string
	Text      string // combined search text
}

func (s *Service) buildSearchText(p *catalog.Product) string {
	parts := []string{p.Name, p.NameEn, p.Description}
	if p.Tags != "" {
		parts = append(parts, strings.ReplaceAll(p.Tags, ",", " "))
	}
	if p.Scenes != "" {
		parts = append(parts, strings.ReplaceAll(p.Scenes, ",", " "))
	}
	return strings.Join(parts, " ")
}

func (s *Service) IndexProduct(p *catalog.Product) error {
	text := s.buildSearchText(p)
	vec, err := s.embeddingCli.Embed(text)
	if err != nil {
		return fmt.Errorf("search.index_error: %w", err)
	}

	row := map[string]interface{}{
		"id":     int64(p.ID),
		"vector": vec,
		"text":   text,
		"name":   p.Name,
		"name_en": p.NameEn,
	}
	return s.milvusCli.Insert(collectionName, []map[string]interface{}{row})
}

func (s *Service) IndexAll() error {
	if err := s.InitCollection(s.embeddingDim); err != nil {
		return fmt.Errorf("search.init_collection_error: %w", err)
	}

	products, _, err := s.catalogRepo.ListProducts(1, 10000, "", 0)
	if err != nil {
		return err
	}

	indexed := 0
	for i := range products {
		if err := s.IndexProduct(&products[i]); err != nil {
			klog.Infof("[search] failed to index product %d: %v", products[i].ID, err)
			continue
		}
		indexed++
	}
	klog.Infof("[search] indexed %d/%d products", indexed, len(products))
	return nil
}

type SearchResult struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	NameEn    string  `json:"name_en"`
	Score     float64 `json:"score"`
}

func (s *Service) Search(query string, limit int) ([]catalog.Product, error) {
	vec, err := s.embeddingCli.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("search.embed_error: %w", err)
	}

	denseResults, err := s.milvusCli.Search(collectionName, vec, s.recallTopK, []string{"id"})
	if err != nil {
		return nil, fmt.Errorf("search.milvus_error: %w", err)
	}

	// Build combined scores (dense + optional BM25)
	combined := make(map[uint]float64)

	// Normalize dense scores
	if len(denseResults) > 0 {
		maxDist := denseResults[0].Distance
		minDist := denseResults[len(denseResults)-1].Distance
		distRange := maxDist - minDist
		if distRange <= 0 {
			distRange = 1.0
		}
		for _, r := range denseResults {
			normScore := (r.Distance - minDist) / distRange
			combined[uint(r.ID)] = s.denseWeight * normScore
		}
	}

	// BM25 keyword search (silent fallback if unavailable)
	if s.bm25Weight > 0 {
		keywords := tokenizeKeywords(query)
		if len(keywords) > 0 {
			keywordIDs, err := s.milvusCli.QueryByKeyword(collectionName, keywords, s.recallTopK)
			if err == nil {
				for i, id := range keywordIDs {
					posScore := 1.0 - float64(i)/float64(len(keywordIDs))
					if existing, ok := combined[uint(id)]; ok {
						combined[uint(id)] = existing + s.bm25Weight*posScore
					} else {
						combined[uint(id)] = s.bm25Weight * posScore
					}
				}
			}
		}
	}

	// Sort by combined score
	type pair struct {
		id    uint
		score float64
	}
	sorted := make([]pair, 0, len(combined))
	for id, score := range combined {
		sorted = append(sorted, pair{id, score})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].score > sorted[j].score })

	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	ids := make([]uint, len(sorted))
	for i, p := range sorted {
		ids[i] = p.id
	}

	products, err := s.catalogRepo.GetProductsByIDs(ids)
	if err != nil {
		return nil, err
	}

	// Preserve sorted order
	productMap := make(map[uint]catalog.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	ordered := make([]catalog.Product, 0, len(sorted))
	for _, p := range sorted {
		if prod, ok := productMap[p.id]; ok {
			ordered = append(ordered, prod)
		}
	}

	return ordered, nil
}

func tokenizeKeywords(query string) []string {
	parts := strings.Fields(query)
	keywords := make([]string, 0, len(parts))
	for _, p := range parts {
		if len([]rune(p)) >= 2 {
			keywords = append(keywords, p)
		}
	}
	if len(keywords) == 0 {
		keywords = parts
	}
	return keywords
}
