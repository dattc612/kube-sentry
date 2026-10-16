import React from "react";
import { createRoot } from "react-dom/client";
const container = document.getElementById("root");
import App from "./App";

if (container) {
  const root = createRoot(container);
  root.render(
    <React.StrictMode>
      <App></App>
    </React.StrictMode>,
  );
}
