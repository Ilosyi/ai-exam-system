import React from "react";
import ReactDOM from "react-dom/client";
import { ConfigProvider, App as AntApp } from "antd";
import App from "./App";
import "./index.css";
import "antd/dist/reset.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#2754c5",
          colorSuccess: "#1f6b5c",
          colorWarning: "#b97418",
          colorError: "#b53d2a",
          colorInfo: "#2754c5",
          borderRadius: 16,
          borderRadiusLG: 24,
          fontFamily: "\"Avenir Next\", \"PingFang SC\", \"Hiragino Sans GB\", \"Microsoft YaHei\", sans-serif",
          colorText: "#203047",
          colorTextSecondary: "#5d6a7e",
          colorBgBase: "#f6f1e6",
          colorBgContainer: "#fffdf8",
          colorBorderSecondary: "rgba(34, 52, 79, 0.1)",
          boxShadowSecondary: "0 18px 48px rgba(25, 41, 72, 0.12)",
        },
        components: {
          Layout: {
            bodyBg: "transparent",
            siderBg: "transparent",
            headerBg: "transparent",
          },
          Card: {
            borderRadiusLG: 24,
          },
          Table: {
            headerBg: "rgba(237, 230, 213, 0.78)",
            headerColor: "#344255",
            rowHoverBg: "rgba(39, 84, 197, 0.05)",
            borderColor: "rgba(34, 52, 79, 0.08)",
          },
          Button: {
            borderRadius: 999,
            primaryShadow: "0 10px 24px rgba(39, 84, 197, 0.22)",
          },
          Drawer: {
            colorBgElevated: "#fffdf8",
          },
          Modal: {
            contentBg: "#fffdf8",
            headerBg: "#fffdf8",
          },
        },
      }}
    >
      <AntApp>
        <App />
      </AntApp>
    </ConfigProvider>
  </React.StrictMode>,
);
