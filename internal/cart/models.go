package cart

type CartItem struct {
	SKUID       uint   `json:"sku_id"`
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	SKUName     string `json:"sku_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	ImageURL    string `json:"image_url"`
	Stock       int    `json:"stock"`
}

type AddItemReq struct {
	SKUID    uint `json:"sku_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,min=1"`
}

type UpdateQtyReq struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}
