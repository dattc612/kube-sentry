import axios from "axios";

// Create a configured instance
export const api = axios.create({
  baseURL: "/api/v1",
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
		if (error.response && error.response.status === 401) {
      localStorage.removeItem('authToken');
    }
    const serverMessage = error.response?.data?.error || error.message || "An unexpected error occurred";
    return Promise.reject(new Error(serverMessage));
  }
);

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('authToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
