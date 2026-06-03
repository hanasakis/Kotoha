package catalog_test

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedData(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := catalog.NewRepository(database)
	svc := catalog.NewService(repo)

	err := svc.SeedData()
	require.NoError(t, err)

	t.Run("categories", func(t *testing.T) {
		cats, err := repo.ListCategories()
		require.NoError(t, err)
		assert.Len(t, cats, 5)
		assert.Equal(t, "现烤坚果炒货", cats[0].Name)
		assert.Equal(t, "Roasted Nuts", cats[0].NameEn)
		assert.Equal(t, 1, cats[0].SortOrder)
		assert.Equal(t, "冻干水果/果脯", cats[1].Name)
		assert.Equal(t, "香辣肉食/海味", cats[2].Name)
		assert.Equal(t, "糕点膨化", cats[3].Name)
		assert.Equal(t, "茶饮搭配包", cats[4].Name)
	})

	t.Run("products", func(t *testing.T) {
		products, total, err := repo.ListProducts(1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(20), total)
		assert.Len(t, products, 20)

		for _, p := range products {
			assert.NotEmpty(t, p.Name)
			assert.NotEmpty(t, p.NameEn)
			assert.NotEmpty(t, p.Description)
			assert.Greater(t, p.CategoryID, uint(0))
			assert.LessOrEqual(t, p.CategoryID, uint(5))
			assert.True(t, p.IsActive)
		}
	})

	t.Run("skus", func(t *testing.T) {
		products, _, _ := repo.ListProducts(1, 100)
		for _, p := range products {
			require.NotEmpty(t, p.SKUs, "product %d (%s) has no SKUs", p.ID, p.Name)
			for _, sku := range p.SKUs {
				assert.Greater(t, sku.Price, 0)
				assert.Greater(t, sku.Stock, 0)
				assert.NotEmpty(t, sku.Name)
			}
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		err := svc.SeedData()
		require.NoError(t, err)
		_, total, _ := repo.ListProducts(1, 100)
		assert.Equal(t, int64(20), total, "seed should be idempotent")
	})
}

func TestSeedDataProductDetails(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := catalog.NewRepository(database)
	svc := catalog.NewService(repo)
	require.NoError(t, svc.SeedData())

	t.Run("prices_in_cents", func(t *testing.T) {
		products, _, _ := repo.ListProducts(1, 100)
		defaultSKUPrices := map[string]int{
			"海盐黄油扁桃仁": 1990, "蜂蜜烤核桃仁": 2290, "日式芥末腰果": 1990,
			"冻干草莓脆": 1590, "芒果干": 1290, "蓝莓果脯": 1690,
			"灯影牛肉丝": 2590, "碳烤鱿鱼丝": 1890, "麻辣鸭脖": 2290,
			"日式芝士脆饼": 1390, "黑芝麻酥": 1690, "海苔肉松卷": 1290,
			"办公室轻松茶饮包": 1990, "低卡花草茶组合": 1690, "椰子脆片": 1490,
			"混合坚果每日包": 3990, "紫薯山药脆": 1190, "泡椒凤爪": 1890,
			"追剧零食豪华包": 9900, "办公室低卡补给包": 3990,
		}

		for _, p := range products {
			expected, ok := defaultSKUPrices[p.Name]
			require.True(t, ok, "unexpected product: %s", p.Name)
			for _, sku := range p.SKUs {
				if sku.IsDefault {
					assert.Equal(t, expected, sku.Price, "price mismatch for %s", p.Name)
				}
			}
		}
	})
}
