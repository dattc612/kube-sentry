import { LoginResponse } from "../models/auth";
import { api } from "./apiClient";

export const authApi = {
  async login(username: string, password: string): Promise<LoginResponse> {
    const { data } = await api.post<LoginResponse>(
      "login",
      { username, password }
    );

    return data;
  },
};