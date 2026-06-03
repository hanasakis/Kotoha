package cart

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/hanasakis/kotoha/pkg/redis"
)

type Repository struct {
	rdb *goredis.Client
}

func NewRepository(rdb *goredis.Client) *Repository {
	return &Repository{rdb: rdb}
}

func (r *Repository) cartKey(userID uint) string {
	return fmt.Sprintf("cart:%d", userID)
}

func (r *Repository) GetItems(ctx context.Context, userID uint) (map[uint]int, error) {
	raw, err := r.rdb.RDB.HGetAll(ctx, r.cartKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	items := make(map[uint]int, len(raw))
	for k, v := range raw {
		skuID, _ := strconv.ParseUint(k, 10, 64)
		qty, _ := strconv.Atoi(v)
		if skuID > 0 && qty > 0 {
			items[uint(skuID)] = qty
		}
	}
	return items, nil
}

func (r *Repository) SetItem(ctx context.Context, userID, skuID uint, qty int) error {
	return r.rdb.RDB.HSet(ctx, r.cartKey(userID), strconv.FormatUint(uint64(skuID), 10), qty).Err()
}

func (r *Repository) RemoveItem(ctx context.Context, userID, skuID uint) error {
	return r.rdb.RDB.HDel(ctx, r.cartKey(userID), strconv.FormatUint(uint64(skuID), 10)).Err()
}

func (r *Repository) Clear(ctx context.Context, userID uint) error {
	return r.rdb.RDB.Del(ctx, r.cartKey(userID)).Err()
}
