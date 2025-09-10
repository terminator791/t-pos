import React from "react";
import { Link } from "react-router-dom";
import TPosLoginForm from "./common/tpos-login-form";
import useDarkMode from "@/hooks/useDarkMode";

// image import
import Logo from "@/assets/images/logo/logo-c.svg";

const TPosLogin = () => {
  const [isDark] = useDarkMode();
  return (
    <>
      <div className="h-full grid w-full grow grid-cols-1 place-items-center pt-10 2xl:pt-0 ">
        <div className=" max-w-[416px] mx-auto w-full  space-y-6">
          <div className="text-center">
            <div className="h-[72px] w-[72px] mx-auto">
              <Link to="/">
                <img
                  src={Logo}
                  alt=""
                  className=" object-contain object-center h-full"
                />
              </Link>
            </div>
            <div className=" text-2xl font-semibold text-gray-600 dark:text-gray-300 mb-1 mt-5">
              T-POS Admin Login
            </div>
            <div className=" text-gray-500 dark:text-gray-400 text-sm">
              Welcome to Terminal Point of Sale
            </div>
          </div>
          <div className="p-6 auth-box">
            <TPosLoginForm />
            <div className=" text-center text-sm mt-5 space-x-1 rtl:space-x-reverse mb-1  ">
              <span> Need support?</span>
              <span>
                <Link to="/contact" className=" text-indigo-500">
                  Contact Admin
                </Link>
              </span>
            </div>
          </div>
          <div className="mt-8 flex justify-center text-xs text-gray-400  pb-10 2xl:pb-0">
            <a href="#">Privacy Notice</a>
            <div className="mx-3 my-1 w-px bg-gray-200 "></div>
            <a href="#">Terms of Service</a>
          </div>
        </div>
      </div>
    </>
  );
};

export default TPosLogin;