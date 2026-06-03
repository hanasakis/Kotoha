package search

import (
	"fmt"
	"log"
	"strings"

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
	recallTopK   int
}

func NewService(milvusCli *milvus.Client, embeddingCli *embedding.Client, catalogRepo *catalog.Repository, denseWeight float64, recallTopK int) *Service {
	return &Service{
		milvusCli:    milvusCli,
		embeddingCli: embeddingCli,
		catalogRepo:  catalogRepo,
		denseWeight:  denseWeight,
		recallTopK:   recallTopK,
	}
}

func (s *Service) InitCollection(dim int) error {
	has, _ := s.milvusCli.HasCollection(collectionName)
	if has {
		return nil
	}
	log.Printf("[search] creating Milvus collection: %s (dim=%d)", collectionName, dim)
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
		"id":     p.ID,
		"vector": vec,
	}
	return s.milvusCli.Insert(collectionName, []map[string]interface{}{row})
}

func (s *Service) IndexAll() error {
	products, _, err := s.catalogRepo.ListProducts(1, 10000)
	if err != nil {
		return err
	}

	for i := range products {
		if err := s.IndexProduct(&products[i]); err != nil {
			log.Printf("[search] failed to index product %d: %v", products[i].ID, err)
			continue
		}
	}
	log.Printf("[search] indexed %d products", len(products))
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

	results, err := s.milvusCli.Search(collectionName, vec, s.recallTopK, []string{"id"})
	if err != nil {
		return nil, fmt.Errorf("search.milvus_error: %w", err)
	}

	ids := make([]uint, 0, len(results))
	for _, r := range results {
		ids = append(ids, uint(r.ID))
	}

	products, err := s.catalogRepo.GetProductsByIDs(ids)
	if err != nil {
		return nil, err
	}

	if len(products) > limit {
		products = products[:limit]
	}
	return products, nil
}
