import path from "node:path";

export const appPort = Number(process.env.PLAYWRIGHT_APP_PORT || "8080");
export const baseURL =
  process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${appPort}`;
export const databaseURL =
  process.env.PLAYWRIGHT_DATABASE_URL ||
  process.env.DATABASE_URL ||
  "postgres://devuser:devpassword123!@127.0.0.1:5132/arkive_dev?sslmode=disable";

export const authDir = path.join(process.cwd(), "playwright", ".auth");
export const authFile = path.join(authDir, "user.json");
export const storageDir = path.join(
  process.cwd(),
  "tmp",
  "playwright",
  "storage",
);

export const e2eAccount = {
  brandName: process.env.PLAYWRIGHT_E2E_BRAND_NAME || "Arkive E2E",
  email: process.env.PLAYWRIGHT_E2E_EMAIL || "numan@gmail.com",
  password: process.env.PLAYWRIGHT_E2E_PASSWORD || "12345678@Aa",
};
