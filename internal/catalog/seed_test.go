package catalog_test

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
)

func TestSeedData(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := catalog.NewRepository(database)
	svc := catalog.NewService(repo, "http://localhost:9000/kotoha-images/products/")

	err := svc.SeedData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("categories", func(t *testing.T) {
		cats, err := repo.ListCategories()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cats) != 6 {
			t.Errorf("got %d categories, want 6", len(cats))
		}
	})

	t.Run("products", func(t *testing.T) {
		products, total, err := repo.ListProducts(1, 100, "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 30 {
			t.Errorf("got total %d, want 30", total)
		}
		if len(products) != 30 {
			t.Errorf("got %d products, want 30", len(products))
		}

		for _, p := range products {
			if p.Name == "" {
				t.Error("expected non-empty Name")
			}
			if p.NameEn == "" {
				t.Error("expected non-empty NameEn")
			}
			if p.Description == "" {
				t.Error("expected non-empty Description")
			}
			if p.CategoryID == 0 {
				t.Error("expected non-zero CategoryID")
			}
			if p.CategoryID > 6 {
				t.Errorf("got CategoryID %d, want <= 6", p.CategoryID)
			}
			if !p.IsActive {
				t.Error("expected IsActive to be true")
			}
		}
	})

	t.Run("skus", func(t *testing.T) {
		products, _, _ := repo.ListProducts(1, 100, "", 0)
		for _, p := range products {
			if len(p.SKUs) == 0 {
				t.Errorf("product %d (%s) has no SKUs", p.ID, p.Name)
				continue
			}
			for _, sku := range p.SKUs {
				if sku.Price <= 0 {
					t.Errorf("expected positive price for SKU %d", sku.ID)
				}
				if sku.Stock <= 0 {
					t.Errorf("expected positive stock for SKU %d", sku.ID)
				}
				if sku.Name == "" {
					t.Errorf("expected non-empty SKU name for SKU %d", sku.ID)
				}
			}
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		err := svc.SeedData()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, total, _ := repo.ListProducts(1, 100, "", 0)
		if total != 30 {
			t.Errorf("got %d products after second seed, want 30 (seed should be idempotent)", total)
		}
	})
}

func TestSeedDataProductDetails(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := catalog.NewRepository(database)
	svc := catalog.NewService(repo, "http://localhost:9000/kotoha-images/products/")
	if err := svc.SeedData(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("prices_in_cents", func(t *testing.T) {
		products, _, _ := repo.ListProducts(1, 100, "", 0)
		defaultSKUPrices := map[string]int{
			"海盐黄油扁桃仁":        1990,
			"蜂蜜烤核桃仁":         2290,
			"日式芥末腰果":         1990,
			"冻干草莓脆":          1590,
			"芒果干":            1290,
			"蓝莓果脯":           1690,
			"灯影牛肉丝":          2590,
			"碳烤鱿鱼丝":          1890,
			"麻辣鸭脖":           2290,
			"日式芝士脆饼":         1390,
			"黑芝麻酥":           1690,
			"海苔肉松卷":          1290,
			"办公室轻松茶饮包":       1990,
			"低卡花草茶组合":        1690,
			"椰子脆片":           1490,
			"混合坚果每日包":        3990,
			"紫薯山药脆":          1190,
			"泡椒凤爪":           1890,
			"追剧零食豪华包":        9900,
			"办公室低卡补给包":       3990,
			"茉莉鲜奶茶":          1590,
			"桂花乌龙鲜奶茶":        1690,
			"Kotoha限定零食罐":      3990,
			"Kotoha帆布袋":        2990,
			"抹茶拿铁坚果仁":        2490,
			"黑糖珍珠奶茶酥":        1890,
			"玫瑰海盐黑巧克力":       2590,
			"柠檬草姜茶":          1990,
			"Kotoha不锈钢保温杯":     5990,
			"Kotoha限定零食大礼包":   19990,
		}

		for _, p := range products {
			expected, ok := defaultSKUPrices[p.Name]
			if !ok {
				t.Errorf("unexpected product in seed: %s", p.Name)
				continue
			}
			for _, sku := range p.SKUs {
				if sku.IsDefault && sku.Price != expected {
					t.Errorf("price mismatch for %s: got %d, want %d", p.Name, sku.Price, expected)
				}
			}
		}
	})
}
