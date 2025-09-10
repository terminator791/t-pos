import { apiSlice } from "../apiSlice";

export const authApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    registerUser: builder.mutation({
      query: (user) => ({
        url: "auth/register",
        method: "POST",
        body: user,
      }),
    }),
    login: builder.mutation({
      query: (data) => ({
        url: "auth/login",
        method: "POST",
        body: data,
      }),
    }),
    getProfile: builder.query({
      query: () => ({
        url: "auth/profile",
        method: "GET",
      }),
    }),
    getPermissions: builder.query({
      query: () => ({
        url: "auth/permissions",
        method: "GET",
      }),
    }),
    refreshToken: builder.mutation({
      query: () => ({
        url: "auth/refresh",
        method: "POST",
      }),
    }),
    createPin: builder.mutation({
      query: (data) => ({
        url: "auth/pin",
        method: "POST",
        body: data,
      }),
    }),
    updatePin: builder.mutation({
      query: (data) => ({
        url: "auth/pin",
        method: "PUT",
        body: data,
      }),
    }),
    deletePin: builder.mutation({
      query: () => ({
        url: "auth/pin",
        method: "DELETE",
      }),
    }),
  }),
});

export const {
  useRegisterUserMutation,
  useLoginMutation,
  useGetProfileQuery,
  useGetPermissionsQuery,
  useRefreshTokenMutation,
  useCreatePinMutation,
  useUpdatePinMutation,
  useDeletePinMutation,
} = authApi;
