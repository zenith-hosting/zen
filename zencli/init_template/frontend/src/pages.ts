import Home from "./pages/Home";
import User from "./pages/User";

export const pages = {
  Home,
  User
};

export type PageName = keyof typeof pages;
