import { createSlice } from "@reduxjs/toolkit";

const storedUser = JSON.parse(localStorage.getItem("user"));
const storedToken = localStorage.getItem("auth_token");
const storedDomain = localStorage.getItem("domain");

export const authSlice = createSlice({
  name: "auth",
  initialState: {
    user: storedUser || null,
    token: storedToken || null,
    domain: storedDomain || null,
    roles: [],
    permissions: [],
    isAuth: !!(storedUser && storedToken),
  },
  reducers: {
    setUser: (state, action) => {
      state.user = action.payload.user;
      state.token = action.payload.token;
      state.domain = action.payload.domain;
      state.roles = action.payload.roles || [];
      state.permissions = action.payload.permissions || [];
      state.isAuth = true;
    },
    setPermissions: (state, action) => {
      state.permissions = action.payload;
    },
    logOut: (state) => {
      state.user = null;
      state.token = null;
      state.domain = null;
      state.roles = [];
      state.permissions = [];
      state.isAuth = false;
      // Clear localStorage
      localStorage.removeItem("user");
      localStorage.removeItem("auth_token");
      localStorage.removeItem("domain");
    },
    updateUser: (state, action) => {
      state.user = { ...state.user, ...action.payload };
      localStorage.setItem("user", JSON.stringify(state.user));
    },
  },
});

export const { setUser, setPermissions, logOut, updateUser } = authSlice.actions;
export default authSlice.reducer;
