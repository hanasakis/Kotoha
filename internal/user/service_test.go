package user_test

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/internal/user"
	"github.com/hanasakis/kotoha/pkg/db"
)

func setupUserService(t *testing.T) *user.Service {
	t.Helper()
	database := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	t.Cleanup(func() { testutil.CleanTestDB(t, database) })
	return user.NewService(database)
}

func TestProfile(t *testing.T) {
	svc := setupUserService(t)

	t.Run("get_empty_profile", func(t *testing.T) {
		profile, err := svc.GetProfile(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.UserID != 1 {
			t.Errorf("got UserID %d, want 1", profile.UserID)
		}
		if profile.Phone != "" {
			t.Error("expected empty phone for new profile")
		}
	})

	t.Run("upsert_new_profile", func(t *testing.T) {
		err := svc.UpsertProfile(&user.Profile{
			UserID: 2,
			Avatar: "https://example.com/avatar.png",
			Phone:  "13800138000",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		profile, err := svc.GetProfile(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.Avatar != "https://example.com/avatar.png" {
			t.Errorf("got avatar %q", profile.Avatar)
		}
		if profile.Phone != "13800138000" {
			t.Errorf("got phone %q, want 13800138000", profile.Phone)
		}
	})

	t.Run("upsert_existing_profile", func(t *testing.T) {
		err := svc.UpsertProfile(&user.Profile{
			UserID: 2,
			Phone:  "13900139000",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		profile, _ := svc.GetProfile(2)
		if profile.Phone != "13900139000" {
			t.Errorf("got phone %q, want 13900139000 (should be updated)", profile.Phone)
		}
		if profile.Avatar != "https://example.com/avatar.png" {
			t.Errorf("got avatar %q, expected preserved from first upsert", profile.Avatar)
		}
	})
}

func TestAddress(t *testing.T) {
	svc := setupUserService(t)

	t.Run("list_empty", func(t *testing.T) {
		addrs, err := svc.GetAddresses(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(addrs) != 0 {
			t.Errorf("expected empty addresses, got %d", len(addrs))
		}
	})

	t.Run("create_address", func(t *testing.T) {
		err := svc.CreateAddress(&user.Address{
			UserID:   1,
			Name:     "张三",
			Phone:    "13800001111",
			Province: "广东",
			City:     "深圳",
			District: "南山区",
			Detail:   "科技园路1号",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addrs, err := svc.GetAddresses(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(addrs) != 1 {
			t.Fatalf("got %d addresses, want 1", len(addrs))
		}
		if addrs[0].Name != "张三" {
			t.Errorf("got name %q, want 张三", addrs[0].Name)
		}
		if addrs[0].City != "深圳" {
			t.Errorf("got city %q, want 深圳", addrs[0].City)
		}
		if addrs[0].IsDefault {
			t.Error("expected IsDefault=false by default")
		}
		if addrs[0].ID == 0 {
			t.Error("expected non-zero ID after create")
		}
	})

	t.Run("create_default_address_clears_others", func(t *testing.T) {
		err := svc.CreateAddress(&user.Address{
			UserID:    1,
			Name:      "李四",
			Phone:     "13900002222",
			Province:  "北京",
			City:      "朝阳区",
			Detail:    "建国路88号",
			IsDefault: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addrs, _ := svc.GetAddresses(1)
		if len(addrs) != 2 {
			t.Fatalf("got %d addresses, want 2", len(addrs))
		}

		// Second address should be default and first
		if !addrs[0].IsDefault {
			t.Error("expected first address (by ordering) to be default")
		}
		if addrs[0].Name != "李四" {
			t.Errorf("expected 李四 to be first (is_default), got %s", addrs[0].Name)
		}
	})

	t.Run("update_address", func(t *testing.T) {
		addrs, _ := svc.GetAddresses(1)
		addr := addrs[1] // The non-default one (张三)
		addr.Detail = "更新后的地址"
		err := svc.UpdateAddress(&addr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addrs, _ = svc.GetAddresses(1)
		for _, a := range addrs {
			if a.ID == addr.ID && a.Detail != "更新后的地址" {
				t.Errorf("expected updated detail, got %q", a.Detail)
			}
		}
	})

	t.Run("update_address_set_default", func(t *testing.T) {
		addrs, _ := svc.GetAddresses(1)
		nonDefault := addrs[1]
		nonDefault.IsDefault = true
		err := svc.UpdateAddress(&nonDefault)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addrs, _ = svc.GetAddresses(1)
		// The updated address should be first now (is_default)
		if !addrs[0].IsDefault {
			t.Error("expected first address to be default")
		}
		if addrs[0].Name != nonDefault.Name {
			t.Errorf("expected %s to be first (now default)", nonDefault.Name)
		}
	})

	t.Run("delete_address", func(t *testing.T) {
		addrs, _ := svc.GetAddresses(1)
		count := len(addrs)
		err := svc.DeleteAddress(1, addrs[0].ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addrs, _ = svc.GetAddresses(1)
		if len(addrs) != count-1 {
			t.Errorf("got %d addresses after delete, want %d", len(addrs), count-1)
		}
	})

	t.Run("delete_nonexistent", func(t *testing.T) {
		err := svc.DeleteAddress(1, 99999)
		if err != nil {
			t.Fatalf("unexpected error (deleting non-existent should be no-op): %v", err)
		}
	})

	t.Run("delete_wrong_user", func(t *testing.T) {
		// Create address for user 10
		svc.CreateAddress(&user.Address{UserID: 10, Name: "test", Phone: "123"})
		addrs, _ := svc.GetAddresses(10)
		addrID := addrs[0].ID
		// Try deleting user 10's address as user 11
		err := svc.DeleteAddress(11, addrID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Address should still exist for user 10
		addrs, _ = svc.GetAddresses(10)
		found := false
		for _, a := range addrs {
			if a.ID == addrID {
				found = true
				break
			}
		}
		if !found {
			t.Error("address should still exist for user 10 (wrong user delete shouldn't affect)")
		}
	})
}

func TestPreference(t *testing.T) {
	svc := setupUserService(t)

	t.Run("get_empty_preference", func(t *testing.T) {
		pref, err := svc.GetPreference(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pref.UserID != 1 {
			t.Errorf("got UserID %d, want 1", pref.UserID)
		}
		if pref.TastePrefs != "" {
			t.Error("expected empty taste prefs for new user")
		}
	})

	t.Run("upsert_new_preference", func(t *testing.T) {
		err := svc.UpsertPreference(&user.Preference{
			UserID:        1,
			DietaryLimits: "素食",
			TastePrefs:    "辣味,甜味",
			ScenePrefs:    "追剧,办公",
			Allergens:     "花生",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pref, err := svc.GetPreference(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pref.DietaryLimits != "素食" {
			t.Errorf("got DietaryLimits %q, want 素食", pref.DietaryLimits)
		}
		if pref.TastePrefs != "辣味,甜味" {
			t.Errorf("got TastePrefs %q, want 辣味,甜味", pref.TastePrefs)
		}
		if pref.ScenePrefs != "追剧,办公" {
			t.Errorf("got ScenePrefs %q, want 追剧,办公", pref.ScenePrefs)
		}
		if pref.Allergens != "花生" {
			t.Errorf("got Allergens %q, want 花生", pref.Allergens)
		}
	})

	t.Run("upsert_existing_preference", func(t *testing.T) {
		err := svc.UpsertPreference(&user.Preference{
			UserID:    1,
			TastePrefs: "酸味",
			Allergens:  "牛奶,花生",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pref, _ := svc.GetPreference(1)
		if pref.TastePrefs != "酸味" {
			t.Errorf("got TastePrefs %q, want 酸味 (updated)", pref.TastePrefs)
		}
		if pref.Allergens != "牛奶,花生" {
			t.Errorf("got Allergens %q, want 牛奶,花生 (updated)", pref.Allergens)
		}
		// These should be preserved from the first upsert
		if pref.DietaryLimits != "素食" {
			t.Errorf("got DietaryLimits %q, want 素食 (preserved)", pref.DietaryLimits)
		}
		if pref.ScenePrefs != "追剧,办公" {
			t.Errorf("got ScenePrefs %q, want 追剧,办公 (preserved)", pref.ScenePrefs)
		}
	})
}
