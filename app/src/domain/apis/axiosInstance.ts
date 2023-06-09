import axios, { AxiosRequestConfig } from 'axios';

export const ErrorUnExpectedResponse = 'Sorry, occurs unexpected errors. Please try agin.';

/**fetch api in csrf-safe */
export const tokenizeFetch = async <T>(config: AxiosRequestConfig) => {
  const instance = axios.create({
    baseURL: '/api',
  });
  const res = await instance.get('/safe');
  const token = res.headers['x-csrf-token'];
  config.headers = {
    ...config.headers,
    'X-CSRF-Token': token,
  };
  return await instance.request<T>(config);
};

/**doesnot use csrftoken */
export const normalFetch = async <T>(config: AxiosRequestConfig) => {
  const instance = axios.create({
    baseURL: '/api',
  });
  return await instance.request<T>(config);
};
