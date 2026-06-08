package search

import (
	"strings"
	"testing"

	"github.com/hanasakis/kotoha/internal/catalog"
)

func TestBuildSearchText(t *testing.T) {
	t.Run("full_product", func(t *testing.T) {
		p := &catalog.Product{
			Name:        "麻辣花生",
			NameEn:      "Spicy Peanuts",
			Description: "香脆可口的麻辣花生，四川风味",
			Tags:        "辣味,花生,零食",
			Scenes:      "聚会,下酒",
		}
		text := (&Service{}).buildSearchText(p)
		if !strings.Contains(text, "麻辣花生") {
			t.Errorf("expected text to contain product name, got: %s", text)
		}
		if !strings.Contains(text, "Spicy Peanuts") {
			t.Errorf("expected text to contain English name, got: %s", text)
		}
		if !strings.Contains(text, "香脆可口的麻辣花生，四川风味") {
			t.Errorf("expected text to contain description, got: %s", text)
		}
		if !strings.Contains(text, "辣味 花生 零食") {
			t.Errorf("expected tags with commas replaced by spaces, got: %s", text)
		}
		if !strings.Contains(text, "聚会 下酒") {
			t.Errorf("expected scenes with commas replaced by spaces, got: %s", text)
		}
	})

	t.Run("product_without_tags_and_scenes", func(t *testing.T) {
		p := &catalog.Product{
			Name:        "原味瓜子",
			NameEn:      "Sunflower Seeds",
			Description: "精选大颗瓜子",
		}
		text := (&Service{}).buildSearchText(p)
		if text != "原味瓜子 Sunflower Seeds 精选大颗瓜子" {
			t.Errorf("got %q, want '原味瓜子 Sunflower Seeds 精选大颗瓜子'", text)
		}
	})

	t.Run("product_with_empty_fields", func(t *testing.T) {
		p := &catalog.Product{
			Name:        "Test",
			Description: "Desc",
		}
		text := (&Service{}).buildSearchText(p)
		if text != "Test  Desc" {
			t.Errorf("got %q, want 'Test  Desc'", text)
		}
	})

	t.Run("product_with_only_name", func(t *testing.T) {
		p := &catalog.Product{Name: "单品"}
		text := (&Service{}).buildSearchText(p)
		if text != "单品  " {
			t.Errorf("got %q, want '单品  '", text)
		}
	})
}

func TestInitCollection_SkipIfNoMilvus(t *testing.T) {
	t.Skip("requires Milvus running locally")
}

func TestIndexAndSearch_SkipIfNoMilvus(t *testing.T) {
	t.Skip("requires Milvus and Ollama embedding service running locally")
}

func TestService_New(t *testing.T) {
	svc := NewService(nil, nil, nil, 0.7, 0.3, 10, 768)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.denseWeight != 0.7 {
		t.Errorf("got denseWeight %f, want 0.7", svc.denseWeight)
	}
	if svc.recallTopK != 10 {
		t.Errorf("got recallTopK %d, want 10", svc.recallTopK)
	}
	if svc.embeddingDim != 768 {
		t.Errorf("got embeddingDim %d, want 768", svc.embeddingDim)
	}
}

func TestCollectionName(t *testing.T) {
	if collectionName != "kotoha_products" {
		t.Errorf("got collectionName %q, want kotoha_products", collectionName)
	}
}
