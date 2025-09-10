import axios from 'axios';
import { toast } from 'react-toastify';

// Create axios instance
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user');
      localStorage.removeItem('domain');
      toast.error('Session expired. Please login again.');
      window.location.href = '/';
      return Promise.reject(error);
    }
    
    if (error.response?.status === 403) {
      toast.error('You do not have permission to perform this action.');
    }
    
    if (error.response?.status >= 500) {
      toast.error('Server error. Please try again later.');
    }
    
    return Promise.reject(error);
  }
);

export default api;