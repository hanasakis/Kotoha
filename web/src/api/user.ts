import apiClient from './client';
import type { Profile, Address, Preference, UpdateProfileInput, CreateAddressInput, UpdatePreferenceInput } from '../types/user';

export async function getProfile(): Promise<Profile> {
  const { data } = await apiClient.get('/profile');
  return data.data;
}

export async function updateProfile(input: UpdateProfileInput): Promise<Profile> {
  const { data } = await apiClient.put('/profile', input);
  return data.data;
}

export async function listAddresses(): Promise<Address[]> {
  const { data } = await apiClient.get('/addresses');
  return data.data;
}

export async function createAddress(input: CreateAddressInput): Promise<Address> {
  const { data } = await apiClient.post('/addresses', input);
  return data.data;
}

export async function updateAddress(id: number, input: CreateAddressInput): Promise<Address> {
  const { data } = await apiClient.put(`/addresses/${id}`, input);
  return data.data;
}

export async function deleteAddress(id: number): Promise<void> {
  await apiClient.delete(`/addresses/${id}`);
}

export async function getPreference(): Promise<Preference> {
  const { data } = await apiClient.get('/preferences');
  return data.data;
}

export async function updatePreference(input: UpdatePreferenceInput): Promise<Preference> {
  const { data } = await apiClient.put('/preferences', input);
  return data.data;
}
