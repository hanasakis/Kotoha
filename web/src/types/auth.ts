export interface User {
  id: number;
  email: string;
  nickname: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResult {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  password: string;
  nickname?: string;
}

export interface ForgotPasswordInput {
  email: string;
}

export interface ResetPasswordInput {
  token: string;
  new_password: string;
}

export interface DeleteAccountInput {
  password: string;
}
