import React, { useRef, useEffect, useState } from "react";
import SidebarLogo from "../sidebar/Logo";
import MainNavmenu from "./MainNavmenu";
import { mainMenuItems } from "@/mocks/main-menu-data";
import SimpleBar from "simplebar-react";
import useSidebar from "@/hooks/useSidebar";
import useSemiDark from "@/hooks/useSemiDark";
import clsx from "clsx";

const MainSidebar = () => {
  const scrollableNodeRef = useRef();
  const [scroll, setScroll] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      if (scrollableNodeRef.current.scrollTop > 0) {
        setScroll(true);
      } else {
        setScroll(false);
      }
    };
    scrollableNodeRef.current.addEventListener("scroll", handleScroll);
  }, [scrollableNodeRef]);

  const [collapsed, setMenuCollapsed] = useSidebar();
  const [menuHover, setMenuHover] = useState(false);

  // semi dark option
  const [isSemiDark] = useSemiDark();

  return (
    <div className={isSemiDark ? "dark" : ""}>
      <div
        className={clsx(
          " sidebar-wrapper bg-white dark:bg-gray-800     shadow-base  ",
          {
            "w-[72px] close_sidebar": collapsed,
            "w-[280px]": !collapsed,
            "sidebar-hovered": menuHover,
          }
        )}
        onMouseEnter={() => {
          setMenuHover(true);
        }}
        onMouseLeave={() => {
          setMenuHover(false);
        }}
      >
        <SidebarLogo menuHover={menuHover} />
        <div
          className={`h-[60px]  absolute top-[80px] nav-shadow z-[1] w-full transition-all duration-200 pointer-events-none ${
            scroll ? " opacity-100" : " opacity-0"
          }`}
        ></div>

        <SimpleBar
          className="sidebar-menu  h-[calc(100%-80px)]"
          scrollableNodeProps={{ ref: scrollableNodeRef }}
        >
          <MainNavmenu menus={mainMenuItems} />
        </SimpleBar>
      </div>
    </div>
  );
};

export default MainSidebar;