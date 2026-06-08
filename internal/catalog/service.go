package catalog

import (
	"fmt"

	klog "github.com/hanasakis/kotoha/pkg/log"
)

type Service struct {
	repo       *Repository
	imgBaseURL string
}

func NewService(repo *Repository, imgBaseURL string) *Service {
	return &Service{repo: repo, imgBaseURL: imgBaseURL}
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

func (s *Service) ListProducts(page, pageSize int, keyword string, categoryID uint) ([]Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.repo.ListProducts(page, pageSize, keyword, categoryID)
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
	_, total, _ := s.repo.ListProducts(1, 1, "", 0)
	if total > 0 {
		// Update image URLs on existing products
		if err := s.updateProductImages(); err != nil {
			klog.Warnf(" Failed to update product images: %v", err)
		}
		// Add new products and categories if they don't exist yet
		return s.ensureNewProducts()
	}

	categories := []Category{
		{Name: "现烤坚果炒货", NameEn: "Roasted Nuts", SortOrder: 1},
		{Name: "冻干水果/果脯", NameEn: "Freeze-Dried Fruits", SortOrder: 2},
		{Name: "香辣肉食/海味", NameEn: "Spicy Meat & Seafood", SortOrder: 3},
		{Name: "糕点膨化", NameEn: "Pastries & Puffs", SortOrder: 4},
		{Name: "茶饮搭配包", NameEn: "Tea Pairing Pack", SortOrder: 5},
		{Name: "Kotoha 品牌周边", NameEn: "Kotoha Merchandise", SortOrder: 6},
	}

	for i := range categories {
		if err := s.repo.CreateCategory(&categories[i]); err != nil {
			return err
		}
	}

	products := []Product{
		{CategoryID: 1, Name: "海盐黄油扁桃仁", NameEn: "Sea Salt Butter Almonds", Description: "低温烘焙，海盐黄油调味，办公室下午茶必备", DescriptionEn: "Low-temp roasted with sea salt butter, perfect office snack", Tags: "坚果,烘焙,下午茶", Scenes: "office,binge-watch", Taste: "salty", Allergens: "nut", WeightGram: 120, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "01-sea-salt-butter-almonds.png",
			SKUs: []SKU{{Name: "120g装", Price: 1990, Stock: 500, IsDefault: true}, {Name: "240g装", Price: 3590, Stock: 300}}},
		{CategoryID: 1, Name: "蜂蜜烤核桃仁", NameEn: "Honey Roasted Walnuts", Description: "云南纸皮核桃，蜂蜜轻烤，甜而不腻", DescriptionEn: "Yunnan paper-shell walnuts with honey light roast", Tags: "坚果,蜂蜜,养生", Scenes: "office,gift", Taste: "sweet", Allergens: "nut", WeightGram: 150, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "02-honey-roasted-walnuts.png",
			SKUs: []SKU{{Name: "150g装", Price: 2290, Stock: 400, IsDefault: true}}},
		{CategoryID: 1, Name: "日式芥末腰果", NameEn: "Japanese Wasabi Cashews", Description: "微辣芥末调味，追剧下酒最佳选择", DescriptionEn: "Mild wasabi seasoning, great with drinks", Tags: "坚果,芥末,零食", Scenes: "binge-watch", Taste: "mild_spicy", Allergens: "nut", WeightGram: 100, ShelfDays: 150, IsActive: true, ImageURL: s.imgBaseURL + "03-wasabi-cashews.png",
			SKUs: []SKU{{Name: "100g装", Price: 1990, Stock: 350, IsDefault: true}}},
		{CategoryID: 2, Name: "冻干草莓脆", NameEn: "Freeze-Dried Strawberry Crisps", Description: "整颗草莓冻干，酥脆酸甜，低卡零添加", DescriptionEn: "Whole strawberry freeze-dried, crispy & sweet, zero additives", Tags: "水果,低卡,健康", Scenes: "office,fitness,kids", Taste: "sweet", Allergens: "", WeightGram: 30, ShelfDays: 365, IsActive: true, ImageURL: s.imgBaseURL + "04-freeze-dried-strawberry.png",
			SKUs: []SKU{{Name: "30g装", Price: 1590, Stock: 600, IsDefault: true}, {Name: "3包装", Price: 3990, Stock: 200}}},
		{CategoryID: 2, Name: "芒果干", NameEn: "Dried Mango", Description: "东南亚芒果鲜切，自然晾晒，厚切有嚼劲", DescriptionEn: "SE Asian mango, sun-dried, thick cut and chewy", Tags: "水果,芒果,酸甜", Scenes: "office,dorm,kids", Taste: "sweet", Allergens: "", WeightGram: 100, ShelfDays: 270, IsActive: true, ImageURL: s.imgBaseURL + "05-dried-mango.png",
			SKUs: []SKU{{Name: "100g装", Price: 1290, Stock: 450, IsDefault: true}, {Name: "200g装", Price: 2290, Stock: 250}}},
		{CategoryID: 2, Name: "蓝莓果脯", NameEn: "Dried Blueberries", Description: "进口蓝莓低温烘干，保留花青素，护眼零食", DescriptionEn: "Imported blueberries oven-dried, retains antioxidants", Tags: "水果,蓝莓,护眼", Scenes: "office,fitness", Taste: "sweet", Allergens: "", WeightGram: 80, ShelfDays: 270, IsActive: true, ImageURL: s.imgBaseURL + "06-dried-blueberries.png",
			SKUs: []SKU{{Name: "80g装", Price: 1690, Stock: 380, IsDefault: true}}},
		{CategoryID: 3, Name: "灯影牛肉丝", NameEn: "Spicy Beef Jerky Shreds", Description: "四川风味，香辣手撕，独立小包装，办公室解馋", DescriptionEn: "Sichuan-style spicy shredded beef, individual packs", Tags: "肉食,香辣,四川", Scenes: "office,binge-watch", Taste: "hot_spicy", Allergens: "", WeightGram: 120, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "07-spicy-beef-jerky-shreds.png",
			SKUs: []SKU{{Name: "120g装", Price: 2590, Stock: 300, IsDefault: true}}},
		{CategoryID: 3, Name: "碳烤鱿鱼丝", NameEn: "Grilled Dried Squid Shreds", Description: "深海鱿鱼碳烤拉丝，鲜香有嚼劲，高蛋白低脂", DescriptionEn: "Deep-sea squid charcoal-grilled, high protein low fat", Tags: "海鲜,高蛋白,低脂", Scenes: "binge-watch,fitness", Taste: "salty", Allergens: "seafood", WeightGram: 80, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "08-grilled-squid-shreds.png",
			SKUs: []SKU{{Name: "80g装", Price: 1890, Stock: 280, IsDefault: true}}},
		{CategoryID: 3, Name: "麻辣鸭脖", NameEn: "Spicy Duck Neck", Description: "精选鸭脖，麻辣鲜香，追剧夜宵首选", DescriptionEn: "Premium duck neck, spicy & savory, perfect late-night snack", Tags: "肉食,麻辣,宵夜", Scenes: "binge-watch,dorm", Taste: "hot_spicy", Allergens: "", WeightGram: 150, ShelfDays: 150, IsActive: true, ImageURL: s.imgBaseURL + "09-spicy-duck-neck.png",
			SKUs: []SKU{{Name: "150g装", Price: 2290, Stock: 320, IsDefault: true}}},
		{CategoryID: 4, Name: "日式芝士脆饼", NameEn: "Japanese Cheese Crackers", Description: "北海道芝士调味，酥脆咸香，独立包装", DescriptionEn: "Hokkaido cheese flavored, crispy & savory, individual packs", Tags: "糕点,芝士,酥脆", Scenes: "office,kids", Taste: "salty", Allergens: "dairy", WeightGram: 100, ShelfDays: 120, IsActive: true, ImageURL: s.imgBaseURL + "10-japanese-cheese-crackers.png",
			SKUs: []SKU{{Name: "100g装", Price: 1390, Stock: 420, IsDefault: true}}},
		{CategoryID: 4, Name: "黑芝麻酥", NameEn: "Black Sesame Crisps", Description: "传统手工黑芝麻酥，无添加糖，老人小孩都爱", DescriptionEn: "Traditional black sesame crisps, no added sugar, all ages", Tags: "糕点,芝麻,无糖", Scenes: "office,gift,kids", Taste: "sweet", Allergens: "", WeightGram: 150, ShelfDays: 120, IsActive: true, ImageURL: s.imgBaseURL + "11-black-sesame-crisps.png",
			SKUs: []SKU{{Name: "150g装", Price: 1690, Stock: 350, IsDefault: true}}},
		{CategoryID: 4, Name: "海苔肉松卷", NameEn: "Seaweed Pork Floss Rolls", Description: "海苔裹肉松，一口酥脆，宿舍必备零食", DescriptionEn: "Seaweed wrapped pork floss, crispy bite-size, dorm essential", Tags: "膨化,肉松,海苔", Scenes: "dorm,binge-watch,kids", Taste: "salty", Allergens: "seafood", WeightGram: 90, ShelfDays: 90, IsActive: true, ImageURL: s.imgBaseURL + "12-seaweed-pork-floss-rolls.png",
			SKUs: []SKU{{Name: "90g装", Price: 1290, Stock: 480, IsDefault: true}}},
		{CategoryID: 5, Name: "办公室轻松茶饮包", NameEn: "Office Easy Tea Pack", Description: "茉莉花茶+绿茶+玫瑰花茶组合，配坚果零食刚刚好", DescriptionEn: "Jasmine + green + rose tea combo, perfect with nuts", Tags: "茶饮,组合,办公室", Scenes: "office,gift", Taste: "low_sugar", Allergens: "", WeightGram: 60, ShelfDays: 365, IsActive: true, ImageURL: s.imgBaseURL + "13-office-tea-pack.png",
			SKUs: []SKU{{Name: "60g体验装", Price: 1990, Stock: 200, IsDefault: true}, {Name: "120g分享装", Price: 3590, Stock: 150}}},
		{CategoryID: 5, Name: "低卡花草茶组合", NameEn: "Low-Cal Herbal Tea Set", Description: "菊花+洛神花+柠檬草，0卡0糖，搭配果干绝配", DescriptionEn: "Chrysanthemum + hibiscus + lemongrass, 0 cal, pairs with dried fruit", Tags: "茶饮,低卡,花茶", Scenes: "office,fitness", Taste: "low_sugar", Allergens: "", WeightGram: 50, ShelfDays: 365, IsActive: true, ImageURL: s.imgBaseURL + "14-herbal-tea-set.png",
			SKUs: []SKU{{Name: "50g装", Price: 1690, Stock: 260, IsDefault: true}}},
		{CategoryID: 2, Name: "椰子脆片", NameEn: "Coconut Chips", Description: "泰国椰子鲜切烘焙，自然甜香，高纤维健康零食", DescriptionEn: "Thai coconut fresh-cut baked, natural sweetness, high fiber", Tags: "水果,椰子,高纤维", Scenes: "office,kids,fitness", Taste: "sweet", Allergens: "", WeightGram: 80, ShelfDays: 240, IsActive: true, ImageURL: s.imgBaseURL + "15-coconut-chips.png",
			SKUs: []SKU{{Name: "80g装", Price: 1490, Stock: 340, IsDefault: true}}},
		{CategoryID: 1, Name: "混合坚果每日包", NameEn: "Daily Mixed Nuts Pack", Description: "杏仁+腰果+核桃+夏威夷果，每日一包营养均衡", DescriptionEn: "Almonds + cashews + walnuts + macadamias, daily nutrition", Tags: "坚果,混合,每日", Scenes: "office,fitness,gift", Taste: "salty", Allergens: "nut", WeightGram: 175, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "16-daily-mixed-nuts.png",
			SKUs: []SKU{{Name: "175g周装(7包)", Price: 3990, Stock: 250, IsDefault: true}, {Name: "750g月装(30包)", Price: 14990, Stock: 80}}},
		{CategoryID: 4, Name: "紫薯山药脆", NameEn: "Purple Yam Crisps", Description: "紫薯+山药低温油炸，低脂非膨化，儿童健康零食", DescriptionEn: "Purple yam + Chinese yam, low-fat non-puffed, kid-friendly", Tags: "膨化,紫薯,低脂,儿童", Scenes: "kids,dorm,office", Taste: "sweet", Allergens: "", WeightGram: 100, ShelfDays: 120, IsActive: true, ImageURL: s.imgBaseURL + "17-purple-yam-crisps.png",
			SKUs: []SKU{{Name: "100g装", Price: 1190, Stock: 400, IsDefault: true}}},
		{CategoryID: 3, Name: "泡椒凤爪", NameEn: "Pickled Chicken Feet", Description: "四川泡椒风味，酸辣开胃，追剧必备", DescriptionEn: "Sichuan pickled pepper flavor, sour-spicy appetizer", Tags: "肉食,泡椒,开胃", Scenes: "binge-watch,dorm", Taste: "hot_spicy", Allergens: "", WeightGram: 130, ShelfDays: 150, IsActive: true, ImageURL: s.imgBaseURL + "18-pickled-chicken-feet.png",
			SKUs: []SKU{{Name: "130g装", Price: 1890, Stock: 310, IsDefault: true}}},
		{CategoryID: 5, Name: "追剧零食豪华包", NameEn: "Binge-Watch Snack Bundle", Description: "芝士脆饼+海苔肉松卷+麻辣鸭脖+冻干草莓，99元追剧一整套", DescriptionEn: "Cheese crackers + seaweed rolls + duck neck + strawberry crisps, all-in-one", Tags: "组合,追剧,超值", Scenes: "binge-watch,gift", Taste: "mixed", Allergens: "dairy,seafood", WeightGram: 420, ShelfDays: 90, IsActive: true, ImageURL: s.imgBaseURL + "19-binge-watch-bundle.png",
			SKUs: []SKU{{Name: "420g超值装", Price: 9900, Stock: 150, IsDefault: true}}},
		{CategoryID: 5, Name: "办公室低卡补给包", NameEn: "Office Low-Cal Supply Pack", Description: "蓝莓果脯+椰子脆片+花草茶组合，全天低卡补给", DescriptionEn: "Blueberries + coconut chips + herbal tea, all-day low-cal supply", Tags: "组合,低卡,办公室", Scenes: "office,fitness", Taste: "low_sugar", Allergens: "", WeightGram: 210, ShelfDays: 240, IsActive: true, ImageURL: s.imgBaseURL + "20-office-low-cal-pack.png",
			SKUs: []SKU{{Name: "210g组合装", Price: 3990, Stock: 180, IsDefault: true}}},
		// New brand-themed products — milk tea, merchandise, snacks
		{CategoryID: 5, Name: "茉莉鲜奶茶", NameEn: "Jasmine Fresh Milk Tea", Description: "茉莉花茶为底，鲜牛乳调配，清爽不腻，办公室下午茶新宠", DescriptionEn: "Jasmine tea base with fresh milk, refreshing and creamy, new office favorite", Tags: "茶饮,奶茶,鲜奶", Scenes: "office,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 280, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "21-jasmine-milk-tea.png",
			SKUs: []SKU{{Name: "280ml瓶装", Price: 1590, Stock: 300, IsDefault: true}, {Name: "6瓶装", Price: 7990, Stock: 100}}},
		{CategoryID: 5, Name: "桂花乌龙鲜奶茶", NameEn: "Osmanthus Oolong Milk Tea", Description: "桂花乌龙茶底+鲜牛乳，花香茶韵交融，秋季限定款", DescriptionEn: "Osmanthus oolong base with fresh milk, floral and tea harmony, autumn limited", Tags: "茶饮,奶茶,桂花,限定", Scenes: "office,gift,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 280, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "22-osmanthus-oolong-milk-tea.png",
			SKUs: []SKU{{Name: "280ml瓶装", Price: 1690, Stock: 250, IsDefault: true}, {Name: "6瓶装", Price: 8590, Stock: 80}}},
		{CategoryID: 6, Name: "Kotoha限定零食罐", NameEn: "Kotoha Limited Snack Jar", Description: "竹盖玻璃罐，日式简约设计，可重复使用，收纳零食或茶叶都很美丽", DescriptionEn: "Glass jar with bamboo lid, Japanese minimalist design, reusable for snacks or tea", Tags: "周边,收纳,日式,品牌", Scenes: "office,gift,home", Taste: "others", Allergens: "", WeightGram: 350, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "23-kotoha-snack-jar.png",
			SKUs: []SKU{{Name: "500ml标准款", Price: 3990, Stock: 200, IsDefault: true}, {Name: "800ml大号款", Price: 5590, Stock: 120}}},
		{CategoryID: 6, Name: "Kotoha帆布袋", NameEn: "Kotoha Canvas Tote Bag", Description: "天然棉帆布袋，零食主题印花，环保出街必备，可装零食也适合日常通勤", DescriptionEn: "Natural cotton canvas tote with snack-themed print, eco-friendly daily bag", Tags: "周边,帆布袋,环保,品牌", Scenes: "office,home,street", Taste: "others", Allergens: "", WeightGram: 120, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "24-kotoha-canvas-bag.png",
			SKUs: []SKU{{Name: "标准款", Price: 2990, Stock: 300, IsDefault: true}}},
		{CategoryID: 1, Name: "抹茶拿铁坚果仁", NameEn: "Matcha Latte Nuts", Description: "进口坚果裹抹茶拿铁粉，日式茶道风味，高颜值健康零食", DescriptionEn: "Premium nuts coated with matcha latte powder, Japanese tea ceremony flavor", Tags: "坚果,抹茶,日式,高颜值", Scenes: "office,gift,afternoon", Taste: "sweet", Allergens: "nut,dairy", WeightGram: 130, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "25-matcha-latte-nuts.png",
			SKUs: []SKU{{Name: "130g装", Price: 2490, Stock: 280, IsDefault: true}, {Name: "260g装", Price: 4290, Stock: 150}}},
		{CategoryID: 4, Name: "黑糖珍珠奶茶酥", NameEn: "Brown Sugar Bubble Tea Crisp", Description: "黑糖珍珠奶茶口味酥饼，外酥内软，珍珠颗粒口感惊喜，网红同款", DescriptionEn: "Brown sugar bubble tea flavored pastry, crispy outside soft inside, trending snack", Tags: "糕点,奶茶,黑糖,网红", Scenes: "office,dorm,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 120, ShelfDays: 90, IsActive: true, ImageURL: s.imgBaseURL + "26-brown-sugar-bubble-tea-crisp.png",
			SKUs: []SKU{{Name: "120g装", Price: 1890, Stock: 350, IsDefault: true}}},
		{CategoryID: 4, Name: "玫瑰海盐黑巧克力", NameEn: "Rose Sea Salt Dark Chocolate", Description: "70%黑巧配玫瑰花瓣+海盐，微苦回甘，高级感零食，适合送礼", DescriptionEn: "70% dark chocolate with rose petals and sea salt, bittersweet elegance, gift-worthy", Tags: "巧克力,玫瑰,高级,送礼", Scenes: "gift,afternoon,home", Taste: "sweet", Allergens: "dairy", WeightGram: 80, ShelfDays: 240, IsActive: true, ImageURL: s.imgBaseURL + "27-rose-sea-salt-chocolate.png",
			SKUs: []SKU{{Name: "80g装", Price: 2590, Stock: 220, IsDefault: true}, {Name: "160g礼盒装", Price: 4590, Stock: 100}}},
		{CategoryID: 5, Name: "柠檬草姜茶", NameEn: "Lemongrass Ginger Tea", Description: "柠檬草+生姜+蜂蜜，暖身驱寒，独立三角茶包，办公桌常备", DescriptionEn: "Lemongrass + ginger + honey, warming and comforting, pyramid tea bags", Tags: "茶饮,姜茶,暖身,健康", Scenes: "office,fitness,home", Taste: "sweet", Allergens: "", WeightGram: 40, ShelfDays: 365, IsActive: true, ImageURL: s.imgBaseURL + "28-lemongrass-ginger-tea.png",
			SKUs: []SKU{{Name: "40g装(15包)", Price: 1990, Stock: 300, IsDefault: true}, {Name: "80g装(30包)", Price: 3490, Stock: 180}}},
		{CategoryID: 6, Name: "Kotoha不锈钢保温杯", NameEn: "Kotoha Stainless Thermos", Description: "316不锈钢内胆，磨砂暖色外观，12小时保温，配茶滤网，办公室必备", DescriptionEn: "316 stainless steel, matte warm finish, 12h insulation, tea strainer included", Tags: "周边,保温杯,品牌,办公室", Scenes: "office,home,street", Taste: "others", Allergens: "", WeightGram: 300, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "29-kotoha-thermos.png",
			SKUs: []SKU{{Name: "350ml标准款", Price: 5990, Stock: 180, IsDefault: true}, {Name: "500ml大容量款", Price: 7990, Stock: 120}}},
		{CategoryID: 6, Name: "Kotoha限定零食大礼包", NameEn: "Kotoha Limited Gift Box", Description: "精选6款人气零食+2款鲜奶茶+周边帆布袋，和风礼盒包装，送礼自用两相宜", DescriptionEn: "6 top snacks + 2 milk teas + canvas tote in Japanese-style gift box", Tags: "礼盒,限定,送礼,超值,组合", Scenes: "gift,home,office", Taste: "mixed", Allergens: "dairy,nut", WeightGram: 1200, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "30-limited-gift-box.png",
			SKUs: []SKU{{Name: "1200g豪华礼盒", Price: 19990, Stock: 80, IsDefault: true}}},
	}

	for i := range products {
		if err := s.repo.CreateProduct(&products[i]); err != nil {
			return err
		}
	}

	return nil
}

// updateProductImages sets image URLs on already-seeded products
func (s *Service) updateProductImages() error {
	mapping := map[string]string{
		"海盐黄油扁桃仁":       s.imgBaseURL + "01-sea-salt-butter-almonds.png",
		"蜂蜜烤核桃仁":        s.imgBaseURL + "02-honey-roasted-walnuts.png",
		"日式芥末腰果":        s.imgBaseURL + "03-wasabi-cashews.png",
		"冻干草莓脆":         s.imgBaseURL + "04-freeze-dried-strawberry.png",
		"芒果干":           s.imgBaseURL + "05-dried-mango.png",
		"蓝莓果脯":          s.imgBaseURL + "06-dried-blueberries.png",
		"灯影牛肉丝":         s.imgBaseURL + "07-spicy-beef-jerky-shreds.png",
		"碳烤鱿鱼丝":         s.imgBaseURL + "08-grilled-squid-shreds.png",
		"麻辣鸭脖":          s.imgBaseURL + "09-spicy-duck-neck.png",
		"日式芝士脆饼":        s.imgBaseURL + "10-japanese-cheese-crackers.png",
		"黑芝麻酥":          s.imgBaseURL + "11-black-sesame-crisps.png",
		"海苔肉松卷":         s.imgBaseURL + "12-seaweed-pork-floss-rolls.png",
		"办公室轻松茶饮包":      s.imgBaseURL + "13-office-tea-pack.png",
		"低卡花草茶组合":       s.imgBaseURL + "14-herbal-tea-set.png",
		"椰子脆片":          s.imgBaseURL + "15-coconut-chips.png",
		"混合坚果每日包":       s.imgBaseURL + "16-daily-mixed-nuts.png",
		"紫薯山药脆":         s.imgBaseURL + "17-purple-yam-crisps.png",
		"泡椒凤爪":          s.imgBaseURL + "18-pickled-chicken-feet.png",
		"追剧零食豪华包":       s.imgBaseURL + "19-binge-watch-bundle.png",
		"办公室低卡补给包":      s.imgBaseURL + "20-office-low-cal-pack.png",
		// New brand-themed products
		"茉莉鲜奶茶":         s.imgBaseURL + "21-jasmine-milk-tea.png",
		"桂花乌龙鲜奶茶":       s.imgBaseURL + "22-osmanthus-oolong-milk-tea.png",
		"Kotoha限定零食罐":   s.imgBaseURL + "23-kotoha-snack-jar.png",
		"Kotoha帆布袋":      s.imgBaseURL + "24-kotoha-canvas-bag.png",
		"抹茶拿铁坚果仁":       s.imgBaseURL + "25-matcha-latte-nuts.png",
		"黑糖珍珠奶茶酥":       s.imgBaseURL + "26-brown-sugar-bubble-tea-crisp.png",
		"玫瑰海盐黑巧克力":      s.imgBaseURL + "27-rose-sea-salt-chocolate.png",
		"柠檬草姜茶":         s.imgBaseURL + "28-lemongrass-ginger-tea.png",
		"Kotoha不锈钢保温杯":  s.imgBaseURL + "29-kotoha-thermos.png",
		"Kotoha限定零食大礼包":  s.imgBaseURL + "30-limited-gift-box.png",
	}
	return s.repo.UpdateProductImages(mapping)
}

// ensureNewProducts adds new categories and products that don't exist yet
func (s *Service) ensureNewProducts() error {
	// Ensure category 6 exists
	cats, _ := s.repo.ListCategories()
	hasCat6 := false
	for _, c := range cats {
		if c.Name == "Kotoha 品牌周边" {
			hasCat6 = true
			break
		}
	}
	if !hasCat6 {
		cat6 := &Category{Name: "Kotoha 品牌周边", NameEn: "Kotoha Merchandise", SortOrder: 6}
		if err := s.repo.CreateCategory(cat6); err != nil {
			return err
		}
	}

	// Check which new products are missing
	_, total, _ := s.repo.ListProducts(1, 50, "", 0)
	if total >= 30 {
		return nil // All 30 products exist
	}

	newProducts := []Product{
		{CategoryID: 5, Name: "茉莉鲜奶茶", NameEn: "Jasmine Fresh Milk Tea", Description: "茉莉花茶为底，鲜牛乳调配，清爽不腻，办公室下午茶新宠", DescriptionEn: "Jasmine tea base with fresh milk, refreshing and creamy, new office favorite", Tags: "茶饮,奶茶,鲜奶", Scenes: "office,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 280, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "21-jasmine-milk-tea.png",
			SKUs: []SKU{{Name: "280ml瓶装", Price: 1590, Stock: 300, IsDefault: true}, {Name: "6瓶装", Price: 7990, Stock: 100}}},
		{CategoryID: 5, Name: "桂花乌龙鲜奶茶", NameEn: "Osmanthus Oolong Milk Tea", Description: "桂花乌龙茶底+鲜牛乳，花香茶韵交融，秋季限定款", DescriptionEn: "Osmanthus oolong base with fresh milk, floral and tea harmony, autumn limited", Tags: "茶饮,奶茶,桂花,限定", Scenes: "office,gift,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 280, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "22-osmanthus-oolong-milk-tea.png",
			SKUs: []SKU{{Name: "280ml瓶装", Price: 1690, Stock: 250, IsDefault: true}, {Name: "6瓶装", Price: 8590, Stock: 80}}},
		{CategoryID: 6, Name: "Kotoha限定零食罐", NameEn: "Kotoha Limited Snack Jar", Description: "竹盖玻璃罐，日式简约设计，可重复使用，收纳零食或茶叶都很美丽", DescriptionEn: "Glass jar with bamboo lid, Japanese minimalist design, reusable for snacks or tea", Tags: "周边,收纳,日式,品牌", Scenes: "office,gift,home", Taste: "others", Allergens: "", WeightGram: 350, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "23-kotoha-snack-jar.png",
			SKUs: []SKU{{Name: "500ml标准款", Price: 3990, Stock: 200, IsDefault: true}, {Name: "800ml大号款", Price: 5590, Stock: 120}}},
		{CategoryID: 6, Name: "Kotoha帆布袋", NameEn: "Kotoha Canvas Tote Bag", Description: "天然棉帆布袋，零食主题印花，环保出街必备，可装零食也适合日常通勤", DescriptionEn: "Natural cotton canvas tote with snack-themed print, eco-friendly daily bag", Tags: "周边,帆布袋,环保,品牌", Scenes: "office,home,street", Taste: "others", Allergens: "", WeightGram: 120, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "24-kotoha-canvas-bag.png",
			SKUs: []SKU{{Name: "标准款", Price: 2990, Stock: 300, IsDefault: true}}},
		{CategoryID: 1, Name: "抹茶拿铁坚果仁", NameEn: "Matcha Latte Nuts", Description: "进口坚果裹抹茶拿铁粉，日式茶道风味，高颜值健康零食", DescriptionEn: "Premium nuts coated with matcha latte powder, Japanese tea ceremony flavor", Tags: "坚果,抹茶,日式,高颜值", Scenes: "office,gift,afternoon", Taste: "sweet", Allergens: "nut,dairy", WeightGram: 130, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "25-matcha-latte-nuts.png",
			SKUs: []SKU{{Name: "130g装", Price: 2490, Stock: 280, IsDefault: true}, {Name: "260g装", Price: 4290, Stock: 150}}},
		{CategoryID: 4, Name: "黑糖珍珠奶茶酥", NameEn: "Brown Sugar Bubble Tea Crisp", Description: "黑糖珍珠奶茶口味酥饼，外酥内软，珍珠颗粒口感惊喜，网红同款", DescriptionEn: "Brown sugar bubble tea flavored pastry, crispy outside soft inside, trending snack", Tags: "糕点,奶茶,黑糖,网红", Scenes: "office,dorm,afternoon", Taste: "sweet", Allergens: "dairy", WeightGram: 120, ShelfDays: 90, IsActive: true, ImageURL: s.imgBaseURL + "26-brown-sugar-bubble-tea-crisp.png",
			SKUs: []SKU{{Name: "120g装", Price: 1890, Stock: 350, IsDefault: true}}},
		{CategoryID: 4, Name: "玫瑰海盐黑巧克力", NameEn: "Rose Sea Salt Dark Chocolate", Description: "70%黑巧配玫瑰花瓣+海盐，微苦回甘，高级感零食，适合送礼", DescriptionEn: "70% dark chocolate with rose petals and sea salt, bittersweet elegance, gift-worthy", Tags: "巧克力,玫瑰,高级,送礼", Scenes: "gift,afternoon,home", Taste: "sweet", Allergens: "dairy", WeightGram: 80, ShelfDays: 240, IsActive: true, ImageURL: s.imgBaseURL + "27-rose-sea-salt-chocolate.png",
			SKUs: []SKU{{Name: "80g装", Price: 2590, Stock: 220, IsDefault: true}, {Name: "160g礼盒装", Price: 4590, Stock: 100}}},
		{CategoryID: 5, Name: "柠檬草姜茶", NameEn: "Lemongrass Ginger Tea", Description: "柠檬草+生姜+蜂蜜，暖身驱寒，独立三角茶包，办公桌常备", DescriptionEn: "Lemongrass + ginger + honey, warming and comforting, pyramid tea bags", Tags: "茶饮,姜茶,暖身,健康", Scenes: "office,fitness,home", Taste: "sweet", Allergens: "", WeightGram: 40, ShelfDays: 365, IsActive: true, ImageURL: s.imgBaseURL + "28-lemongrass-ginger-tea.png",
			SKUs: []SKU{{Name: "40g装(15包)", Price: 1990, Stock: 300, IsDefault: true}, {Name: "80g装(30包)", Price: 3490, Stock: 180}}},
		{CategoryID: 6, Name: "Kotoha不锈钢保温杯", NameEn: "Kotoha Stainless Thermos", Description: "316不锈钢内胆，磨砂暖色外观，12小时保温，配茶滤网，办公室必备", DescriptionEn: "316 stainless steel, matte warm finish, 12h insulation, tea strainer included", Tags: "周边,保温杯,品牌,办公室", Scenes: "office,home,street", Taste: "others", Allergens: "", WeightGram: 300, ShelfDays: 0, IsActive: true, ImageURL: s.imgBaseURL + "29-kotoha-thermos.png",
			SKUs: []SKU{{Name: "350ml标准款", Price: 5990, Stock: 180, IsDefault: true}, {Name: "500ml大容量款", Price: 7990, Stock: 120}}},
		{CategoryID: 6, Name: "Kotoha限定零食大礼包", NameEn: "Kotoha Limited Gift Box", Description: "精选6款人气零食+2款鲜奶茶+周边帆布袋，和风礼盒包装，送礼自用两相宜", DescriptionEn: "6 top snacks + 2 milk teas + canvas tote in Japanese-style gift box", Tags: "礼盒,限定,送礼,超值,组合", Scenes: "gift,home,office", Taste: "mixed", Allergens: "dairy,nut", WeightGram: 1200, ShelfDays: 180, IsActive: true, ImageURL: s.imgBaseURL + "30-limited-gift-box.png",
			SKUs: []SKU{{Name: "1200g豪华礼盒", Price: 19990, Stock: 80, IsDefault: true}}},
	}

	for i := range newProducts {
		if err := s.repo.CreateProduct(&newProducts[i]); err != nil {
			return err
		}
	}

	return nil
}
