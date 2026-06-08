export interface Profile {
  id: number;
  user_id: number;
  avatar: string;
  phone: string;
  created_at: string;
  updated_at: string;
}

export interface Address {
  id: number;
  user_id: number;
  name: string;
  phone: string;
  province: string;
  city: string;
  district: string;
  detail: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface Preference {
  id: number;
  user_id: number;
  dietary_limits: string;
  taste_prefs: string;
  scene_prefs: string;
  allergens: string;
}

export interface UpdateProfileInput {
  avatar?: string;
  phone?: string;
}

export interface CreateAddressInput {
  name: string;
  phone: string;
  province: string;
  city: string;
  district: string;
  detail: string;
  is_default?: boolean;
}

export interface UpdatePreferenceInput {
  dietary_limits?: string;
  taste_prefs?: string;
  scene_prefs?: string;
  allergens?: string;
}
