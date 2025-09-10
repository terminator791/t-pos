import React, { useState } from "react";
import InputGroup from "@/components/ui/InputGroup";
import Button from "@/components/ui/Button";
import Icon from "@/components/ui/Icon";
import { useForm } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import { useNavigate } from "react-router-dom";
import Checkbox from "@/components/ui/Checkbox";
import { Link } from "react-router-dom";
import { useDispatch } from "react-redux";
import { toast } from "react-toastify";
import { useLoginMutation } from "@/store/api/auth/authApiSlice";
import { setUser } from "@/store/api/auth/authSlice";

const schema = yup
  .object({
    username: yup.string().required("Username is Required"),
    pin: yup
      .string()
      .required("PIN is Required")
      .min(6, "PIN must be 6 digits")
      .max(6, "PIN must be 6 digits"),
  })
  .required();

const TPosLoginForm = () => {
  const [login, { isLoading, isError, error, isSuccess }] = useLoginMutation();
  const dispatch = useDispatch();
  const {
    register,
    reset,
    formState: { errors },
    handleSubmit,
  } = useForm({
    resolver: yupResolver(schema),
    mode: "all",
  });
  const navigate = useNavigate();

  const onSubmit = async (data) => {
    try {
      // Prepare the login data according to backend API format
      const loginData = {
        username: data.username,
        pin: data.pin,
      };

      const response = await login(loginData).unwrap();

      if (response.status === "success" && response.data.token) {
        // Store token and user data
        localStorage.setItem("auth_token", response.data.token);
        localStorage.setItem("user", JSON.stringify(response.data.user));
        localStorage.setItem("domain", response.data.domain || "");
        localStorage.setItem(
          "roles",
          JSON.stringify(response.data.roles || [])
        );

        // Update Redux state
        dispatch(
          setUser({
            user: response.data.user,
            token: response.data.token,
            domain: response.data.domain || "",
            roles: response.data.roles || [],
          })
        );

        // Navigate to main dashboard
        navigate("/main/dashboard");
        toast.success("Login Successful");
        reset();
      } else {
        throw new Error(response.message || "Login failed");
      }
    } catch (error) {
      console.error("Login error:", error);
      const errorMessage =
        error?.data?.message ||
        error?.message ||
        "Login failed. Please check your credentials.";
      toast.error(errorMessage);
    }
  };

  const [checked, setChecked] = useState(false);

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 ">
      <InputGroup
        name="username"
        type="text"
        label="Username"
        placeholder="Enter your username"
        prepend={<Icon icon="ph:user" />}
        register={register}
        error={errors.username}
        merged
        disabled={isLoading}
      />
      <InputGroup
        name="pin"
        label="PIN"
        type="password"
        placeholder="Enter your 6-digit PIN"
        prepend={<Icon icon="ph:lock-simple" />}
        register={register}
        error={errors.pin}
        merged
        disabled={isLoading}
        maxLength={6}
      />

      <div className="flex justify-between">
        <Checkbox
          value={checked}
          onChange={() => setChecked(!checked)}
          label="Remember me"
        />
        <Link
          to="/forgot-password"
          className="text-sm text-gray-400 dark:text-gray-400 hover:text-indigo-500 hover:underline  "
        >
          Forgot PIN?
        </Link>
      </div>

      <Button
        type="submit"
        text="Sign in to T-POS"
        className="btn btn-primary block w-full text-center "
        isLoading={isLoading}
      />
    </form>
  );
};

export default TPosLoginForm;
