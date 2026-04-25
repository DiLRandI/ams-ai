import { expect, test } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";

const apiBaseURL = process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080";
const here = path.dirname(fileURLToPath(import.meta.url));

test("core MVP happy path", async ({ page, request }) => {
  const suffix = Date.now();
  const assetName = `Smoke Warranty Asset ${suffix}`;
  const warrantyExpiry = isoDateOffset(14);

  await page.goto("/login");
  await page.getByLabel("Email").fill("admin@example.com");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();

  await page.getByRole("link", { name: "Asset list" }).click();
  await page.getByRole("link", { name: "Add asset" }).click();
  const categorySelect = page.locator('select[name="categoryId"]');
  await expect(categorySelect).toContainText("IT devices");
  await categorySelect.selectOption({ label: "IT devices" });
  await expect(categorySelect).not.toHaveValue("0");
  await page.getByLabel("Asset name").fill(assetName);
  await page.getByLabel("Brand").fill("SmokeCo");
  await page.getByLabel("Model").fill("E2E-1");
  await page.getByLabel("Serial number / VIN").fill(`SMOKE-${suffix}`);
  await page.getByLabel("Location").fill("Release lab");
  await page.getByLabel("Warranty start").fill(isoDateOffset(0));
  await page.getByLabel("Warranty expiry").fill(warrantyExpiry);
  await page.getByLabel("Warranty notes").fill("E2E warranty visibility");
  const createAssetResponse = page.waitForResponse((response) => {
    return (
      response.request().method() === "POST" &&
      new URL(response.url()).pathname === "/api/assets"
    );
  });
  await page.getByRole("button", { name: "Save asset" }).click();
  const createAsset = await createAssetResponse;
  expect(createAsset.status()).toBe(201);

  await expect(page.getByRole("heading", { name: assetName })).toBeVisible();
  await expect(page.getByText("Expiring soon").first()).toBeVisible();

  await page.getByPlaceholder("Document title").fill("Smoke invoice");
  await page
    .locator('input[name="file"]')
    .setInputFiles(path.join(here, "fixtures", "demo-invoice.pdf"));
  const uploadResponse = page.waitForResponse((response) => {
    return (
      response.request().method() === "POST" &&
      /\/api\/assets\/\d+\/documents$/.test(new URL(response.url()).pathname)
    );
  });
  await page.getByRole("button", { name: /Upload/ }).click();
  const documentPayload = (await (await uploadResponse).json()) as {
    id: number;
  };
  await expect(page.getByText("Smoke invoice")).toBeVisible();

  const unauthorized = await request.get(
    `${apiBaseURL}/api/documents/${documentPayload.id}/download`,
  );
  expect(unauthorized.status()).toBe(401);

  const token = await page.evaluate(() => localStorage.getItem("ams_token"));
  expect(token).toBeTruthy();
  const authorized = await request.get(
    `${apiBaseURL}/api/documents/${documentPayload.id}/download`,
    { headers: { Authorization: `Bearer ${token}` } },
  );
  expect(authorized.status()).toBe(200);
  expect(authorized.headers()["content-type"]).toContain("application/pdf");

  await page.getByRole("link", { name: "Asset list" }).click();
  await page.getByPlaceholder(/Search name/).fill(assetName);
  await page.getByLabel("Warranty state").selectOption("expiring_soon");
  await page.getByLabel("Documents").selectOption("true");
  await expect(page.getByRole("link", { name: assetName })).toBeVisible();
  await page.getByRole("link", { name: assetName }).click();
  await expect(page.getByRole("heading", { name: assetName })).toBeVisible();

  await page.getByRole("link", { name: /Reminders/ }).click();
  await expect(
    page.getByRole("link", { name: `${assetName} warranty expires` }),
  ).toBeVisible();
});

function isoDateOffset(days: number) {
  const date = new Date();
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}
