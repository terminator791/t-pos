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
    email: yup.string().email("Invalid email").required("Email is Required"),
    password: yup.string().required("Password is Required"),
    domain: yup.string().required("Domain/License is Required"),
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
      const response = await login(data).unwrap();

      if (response.status === "success" && response.data.token) {
        // Store token and user data
        localStorage.setItem("auth_token", response.data.token);
        localStorage.setItem("user", JSON.stringify(response.data.user));
        localStorage.setItem("domain", response.data.domain || data.domain);
        
        // Update Redux state
        dispatch(setUser({
          user: response.data.user,
          token: response.data.token,
          domain: response.data.domain || data.domain,
          roles: response.data.roles || []
        }));

        // Navigate to main dashboard
        navigate("/main/dashboard");
        toast.success("Login Successful");
        reset();
      } else {
        throw new Error(response.message || "Login failed");
      }
    } catch (error) {
      console.error("Login error:", error);
      const errorMessage = error?.data?.message || error?.message || "Login failed. Please check your credentials.";
      toast.error(errorMessage);
    }
  };

  const [checked, setChecked] = useState(false);

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 ">
      <InputGroup
        name="email"
        type="email"
        label="Email"
        placeholder="Enter your email"
        prepend="@"
        register={register}
        error={errors.email}
        merged
        disabled={isLoading}
      />
      <InputGroup
        name="password"
        label="Password"
        type="password"
        placeholder="Enter your password"
        prepend={<Icon icon="ph:lock-simple" />}
        register={register}
        error={errors.password}
        merged
        disabled={isLoading}
      />
      <InputGroup
        name="domain"
        type="text"
        label="License/Domain"
        placeholder="Enter license serial or domain"
        prepend={<Icon icon="ph:key" />}
        register={register}
        error={errors.domain}
        merged
        disabled={isLoading}
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
          Forgot Password?
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