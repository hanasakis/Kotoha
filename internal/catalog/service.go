package catalog

import (
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// --- Categories ---

func (s *Service) ListCategories() ([]Category, error) {
	return s.repo.ListCategories()
}

func (s *Service) CreateCategory(name, nameEn string, parentID *uint) (*Category, error) {
	cat := &Category{Name: name, NameEn: nameEn, ParentID: parentID}
	if err := s.repo.CreateCategory(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// --- Products ---

func (s *Service) ListProducts(page, pageSize int) ([]Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.repo.ListProducts(page, pageSize)
}

func (s *Service) GetProduct(id uint) (*Product, error) {
	p, err := s.repo.GetProduct(id)
	if err != nil {
		return nil, fmt.Errorf("product.not_found")
	}
	s.repo.IncrementClickCount(id)
	return p, nil
}

func (s *Service) CreateProduct(p *Product) error {
	return s.repo.CreateProduct(p)
}

func (s *Service) UpdateProduct(p *Product) error {
	return s.repo.UpdateProduct(p)
}

func (s *Service) DeleteProduct(id uint) error {
	return s.repo.DeleteProduct(id)
}

// SeedData inserts demo snack products for Kotoha MVP
func (s *Service) SeedData() error {
	// Check if data already exists
	_, total, _ := s.repo.ListProducts(1, 1)
	if total > 0 {
		return nil
	}

	categories := []Category{
		{Name: "现烤坚果炒货", NameEn: "Roasted Nuts", SortOrder: 1},
		{Name: "冻干水果/果脯", NameEn: "Freeze-Dried Fruits", SortOrder: 2},
		{Name: "香辣肉食/海味", NameEn: "Spicy Meat & Seafood", SortOrder: 3},
		{Name: "糕点膨化", NameEn: "Pastries & Puffs", SortOrder: 4},
		{Name: "茶饮搭配包", NameEn: "Tea Pairing Pack", SortOrder: 5},
	}

	for i := range categories {
		if err := s.repo.CreateCategory(&categories[i]); err != nil {
			return err
		}
	}

	products := []Product{
		{CategoryID: 1, Name: "海盐黄油扁桃仁", NameEn: "Sea Salt Butter Almonds", Description: "低温烘焙，海盐黄油调味，办公室下午茶必备", DescriptionEn: "Low-temp roasted with sea salt butter, perfect office snack", Tags: "坚果,烘焙,下午茶", Scenes: "office,binge-watch", Taste: "salty", Allergens: "nut", WeightGram: 120, ShelfDays: 180, IsActive: true,
			SKUs: []SKU{{Name: "120g装", Price: 1990, Stock: 500, IsDefault: true}, {Name: "240g装", Price: 3590, Stock: 300}}},
		{CategoryID: 1, Name: "蜂蜜烤核桃仁", NameEn: "Honey Roasted Walnuts", Description: "云南纸皮核桃，蜂蜜轻烤，甜而不腻", DescriptionEn: "Yunnan paper-shell walnuts with honey light roast", Tags: "坚果,蜂蜜,养生", Scenes: "office,gift", Taste: "sweet", Allergens: "nut", WeightGram: 150, ShelfDays: 180, IsActive: true,
			SKUs: []SKU{{Name: "150g装", Price: 2290, Stock: 400, IsDefault: true}}},
		{CategoryID: 1, Name: "日式芥末腰果", NameEn: "Japanese Wasabi Cashews", Description: "微辣芥末调味，追剧下酒最佳选择", DescriptionEn: "Mild wasabi seasoning, great with drinks", Tags: "坚果,芥末,零食", Scenes: "binge-watch", Taste: "mild_spicy", Allergens: "nut", WeightGram: 100, ShelfDays: 150, IsActive: true,
			SKUs: []SKU{{Name: "100g装", Price: 1990, Stock: 350, IsDefault: true}}},
		{CategoryID: 2, Name: "冻干草莓脆", NameEn: "Freeze-Dried Strawberry Crisps", Description: "整颗草莓冻干，酥脆酸甜，低卡零添加", DescriptionEn: "Whole strawberry freeze-dried, crispy & sweet, zero additives", Tags: "水果,低卡,健康", Scenes: "office,fitness,kids", Taste: "sweet", Allergens: "", WeightGram: 30, ShelfDays: 365, IsActive: true,
			SKUs: []SKU{{Name: "30g装", Price: 1590, Stock: 600, IsDefault: true}, {Name: "3包装", Price: 3990, Stock: 200}}},
		{CategoryID: 2, Name: "芒果干", NameEn: "Dried Mango", Description: "东南亚芒果鲜切，自然晾晒，厚切有嚼劲", DescriptionEn: "SE Asian mango, sun-dried, thick cut and chewy", Tags: "水果,芒果,酸甜", Scenes: "office,dorm,kids", Taste: "sweet", Allergens: "", WeightGram: 100, ShelfDays: 270, IsActive: true,
			SKUs: []SKU{{Name: "100g装", Price: 1290, Stock: 450, IsDefault: true}, {Name: "200g装", Price: 2290, Stock: 250}}},
		{CategoryID: 2, Name: "蓝莓果脯", NameEn: "Dried Blueberries", Description: "进口蓝莓低温烘干，保留花青素，护眼零食", DescriptionEn: "Imported blueberries oven-dried, retains antioxidants", Tags: "水果,蓝莓,护眼", Scenes: "office,fitness", Taste: "sweet", Allergens: "", WeightGram: 80, ShelfDays: 270, IsActive: true,
			SKUs: []SKU{{Name: "80g装", Price: 1690, Stock: 380, IsDefault: true}}},
		{CategoryID: 3, Name: "灯影牛肉丝", NameEn: "Spicy Beef Jerky Shreds", Description: "四川风味，香辣手撕，独立小包装，办公室解馋", DescriptionEn: "Sichuan-style spicy shredded beef, individual packs", Tags: "肉食,香辣,四川", Scenes: "office,binge-watch", Taste: "hot_spicy", Allergens: "", WeightGram: 120, ShelfDays: 180, IsActive: true,
			SKUs: []SKU{{Name: "120g装", Price: 2590, Stock: 300, IsDefault: true}}},
		{CategoryID: 3, Name: "碳烤鱿鱼丝", NameEn: "Grilled Dried Squid Shreds", Description: "深海鱿鱼碳烤拉丝，鲜香有嚼劲，高蛋白低脂", DescriptionEn: "Deep-sea squid charcoal-grilled, high protein low fat", Tags: "海鲜,高蛋白,低脂", Scenes: "binge-watch,fitness", Taste: "salty", Allergens: "seafood", WeightGram: 80, ShelfDays: 180, IsActive: true,
			SKUs: []SKU{{Name: "80g装", Price: 1890, Stock: 280, IsDefault: true}}},
		{CategoryID: 3, Name: "麻辣鸭脖", NameEn: "Spicy Duck Neck", Description: "精选鸭脖，麻辣鲜香，追剧夜宵首选", DescriptionEn: "Premium duck neck, spicy & savory, perfect late-night snack", Tags: "肉食,麻辣,宵夜", Scenes: "binge-watch,dorm", Taste: "hot_spicy", Allergens: "", WeightGram: 150, ShelfDays: 150, IsActive: true,
			SKUs: []SKU{{Name: "150g装", Price: 2290, Stock: 320, IsDefault: true}}},
		{CategoryID: 4, Name: "日式芝士脆饼", NameEn: "Japanese Cheese Crackers", Description: "北海道芝士调味，酥脆咸香，独立包装", DescriptionEn: "Hokkaido cheese flavored, crispy & savory, individual packs", Tags: "糕点,芝士,酥脆", Scenes: "office,kids", Taste: "salty", Allergens: "dairy", WeightGram: 100, ShelfDays: 120, IsActive: true,
			SKUs: []SKU{{Name: "100g装", Price: 1390, Stock: 420, IsDefault: true}}},
		{CategoryID: 4, Name: "黑芝麻酥", NameEn: "Black Sesame Crisps", Description: "传统手工黑芝麻酥，无添加糖，老人小孩都爱", DescriptionEn: "Traditional black sesame crisps, no added sugar, all ages", Tags: "糕点,芝麻,无糖", Scenes: "office,gift,kids", Taste: "sweet", Allergens: "", WeightGram: 150, ShelfDays: 120, IsActive: true,
			SKUs: []SKU{{Name: "150g装", Price: 1690, Stock: 350, IsDefault: true}}},
		{CategoryID: 4, Name: "海苔肉松卷", NameEn: "Seaweed Pork Floss Rolls", Description: "海苔裹肉松，一口酥脆，宿舍必备零食", DescriptionEn: "Seaweed wrapped pork floss, crispy bite-size, dorm essential", Tags: "膨化,肉松,海苔", Scenes: "dorm,binge-watch,kids", Taste: "salty", Allergens: "seafood", WeightGram: 90, ShelfDays: 90, IsActive: true,
			SKUs: []SKU{{Name: "90g装", Price: 1290, Stock: 480, IsDefault: true}}},
		{CategoryID: 5, Name: "办公室轻松茶饮包", NameEn: "Office Easy Tea Pack", Description: "茉莉花茶+绿茶+玫瑰花茶组合，配坚果零食刚刚好", DescriptionEn: "Jasmine + green + rose tea combo, perfect with nuts", Tags: "茶饮,组合,办公室", Scenes: "office,gift", Taste: "low_sugar", Allergens: "", WeightGram: 60, ShelfDays: 365, IsActive: true,
			SKUs: []SKU{{Name: "60g体验装", Price: 1990, Stock: 200, IsDefault: true}, {Name: "120g分享装", Price: 3590, Stock: 150}}},
		{CategoryID: 5, Name: "低卡花草茶组合", NameEn: "Low-Cal Herbal Tea Set", Description: "菊花+洛神花+柠檬草，0卡0糖，搭配果干绝配", DescriptionEn: "Chrysanthemum + hibiscus + lemongrass, 0 cal, pairs with dried fruit", Tags: "茶饮,低卡,花茶", Scenes: "office,fitness", Taste: "low_sugar", Allergens: "", WeightGram: 50, ShelfDays: 365, IsActive: true,
			SKUs: []SKU{{Name: "50g装", Price: 1690, Stock: 260, IsDefault: true}}},
		{CategoryID: 2, Name: "椰子脆片", NameEn: "Coconut Chips", Description: "泰国椰子鲜切烘焙，自然甜香，高纤维健康零食", DescriptionEn: "Thai coconut fresh-cut baked, natural sweetness, high fiber", Tags: "水果,椰子,高纤维", Scenes: "office,kids,fitness", Taste: "sweet", Allergens: "", WeightGram: 80, ShelfDays: 240, IsActive: true,
			SKUs: []SKU{{Name: "80g装", Price: 1490, Stock: 340, IsDefault: true}}},
		{CategoryID: 1, Name: "混合坚果每日包", NameEn: "Daily Mixed Nuts Pack", Description: "杏仁+腰果+核桃+夏威夷果，每日一包营养均衡", DescriptionEn: "Almonds + cashews + walnuts + macadamias, daily nutrition", Tags: "坚果,混合,每日", Scenes: "office,fitness,gift", Taste: "salty", Allergens: "nut", WeightGram: 175, ShelfDays: 180, IsActive: true,
			SKUs: []SKU{{Name: "175g周装(7包)", Price: 3990, Stock: 250, IsDefault: true}, {Name: "750g月装(30包)", Price: 14990, Stock: 80}}},
		{CategoryID: 4, Name: "紫薯山药脆", NameEn: "Purple Yam Crisps", Description: "紫薯+山药低温油炸，低脂非膨化，儿童健康零食", DescriptionEn: "Purple yam + Chinese yam, low-fat non-puffed, kid-friendly", Tags: "膨化,紫薯,低脂,儿童", Scenes: "kids,dorm,office", Taste: "sweet", Allergens: "", WeightGram: 100, ShelfDays: 120, IsActive: true,
			SKUs: []SKU{{Name: "100g装", Price: 1190, Stock: 400, IsDefault: true}}},
		{CategoryID: 3, Name: "泡椒凤爪", NameEn: "Pickled Chicken Feet", Description: "四川泡椒风味，酸辣开胃，追剧必备", DescriptionEn: "Sichuan pickled pepper flavor, sour-spicy appetizer", Tags: "肉食,泡椒,开胃", Scenes: "binge-watch,dorm", Taste: "hot_spicy", Allergens: "", WeightGram: 130, ShelfDays: 150, IsActive: true,
			SKUs: []SKU{{Name: "130g装", Price: 1890, Stock: 310, IsDefault: true}}},
		{CategoryID: 5, Name: "追剧零食豪华包", NameEn: "Binge-Watch Snack Bundle", Description: "芝士脆饼+海苔肉松卷+麻辣鸭脖+冻干草莓，99元追剧一整套", DescriptionEn: "Cheese crackers + seaweed rolls + duck neck + strawberry crisps, all-in-one", Tags: "组合,追剧,超值", Scenes: "binge-watch,gift", Taste: "mixed", Allergens: "dairy,seafood", WeightGram: 420, ShelfDays: 90, IsActive: true,
			SKUs: []SKU{{Name: "420g超值装", Price: 9900, Stock: 150, IsDefault: true}}},
		{CategoryID: 5, Name: "办公室低卡补给包", NameEn: "Office Low-Cal Supply Pack", Description: "蓝莓果脯+椰子脆片+花草茶组合，全天低卡补给", DescriptionEn: "Blueberries + coconut chips + herbal tea, all-day low-cal supply", Tags: "组合,低卡,办公室", Scenes: "office,fitness", Taste: "low_sugar", Allergens: "", WeightGram: 210, ShelfDays: 240, IsActive: true,
			SKUs: []SKU{{Name: "210g组合装", Price: 3990, Stock: 180, IsDefault: true}}},
	}

	for i := range products {
		if err := s.repo.CreateProduct(&products[i]); err != nil {
			return err
		}
	}

	return nil
}
