import React, { useEffect, useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import Icon from "@/components/ui/Icon";
import { useDispatch } from "react-redux";
import useMobileMenu from "@/hooks/useMobileMenu";

const MainNavmenu = ({ menus }) => {
  const location = useLocation();
  const locationName = location.pathname.replace("/", "");
  const [mobileMenu, setMobileMenu] = useMobileMenu();
  const dispatch = useDispatch();

  useEffect(() => {
    document.title = `T-POS | ${locationName}`;
    if (mobileMenu) {
      setMobileMenu(false);
    }
  }, [location]);

  return (
    <>
      <ul>
        {menus.map((item, i) => (
          <li
            key={i}
            className={`single-menu-item ${
              locationName === item.link ? "menu-item-active" : ""
            }`}
          >
            <NavLink
              to={`/${item.link}`}
              className={({ isActive }) =>
                isActive
                  ? "menu-link menu-link-active"
                  : "menu-link"
              }
            >
              <div className="flex items-center">
                <span className="menu-icon">
                  <Icon icon={item.icon} />
                </span>
                <div className="text-box">{item.title}</div>
              </div>
            </NavLink>
          </li>
        ))}
      </ul>
    </>
  );
};

export default MainNavmenu;