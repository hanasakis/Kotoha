import apiClient from './client';
import type { AuthResult, LoginInput, RegisterInput, ForgotPasswordInput, ResetPasswordInput, DeleteAccountInput } from '../types/auth';

export async function register(input: RegisterInput): Promise<AuthResult> {
  const { data } = await apiClient.post('/auth/register', input);
  return data.data;
}

export async function login(input: LoginInput): Promise<AuthResult> {
  const { data } = await apiClient.post('/auth/login', input);
  return data.data;
}

export async function refreshToken(refreshToken: string): Promise<AuthResult> {
  const { data } = await apiClient.post('/auth/refresh', { refresh_token: refreshToken });
  return data.data;
}

export async function logout(refreshToken: string): Promise<void> {
  await apiClient.post('/auth/logout', { refresh_token: refreshToken });
}

export async function forgotPassword(input: ForgotPasswordInput): Promise<void> {
  await apiClient.post('/auth/forgot-password', input);
}

export async function resetPassword(input: ResetPasswordInput): Promise<void> {
  await apiClient.post('/auth/reset-password', input);
}

export async function deleteAccount(input: DeleteAccountInput): Promise<void> {
  await apiClient.delete('/auth/account', { data: input });
}
