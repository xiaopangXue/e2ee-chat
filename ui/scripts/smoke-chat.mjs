import { chromium } from "playwright";

const tinyPng = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
  "base64",
);

const baseURL = process.env.BASE_URL || "http://127.0.0.1:8080";

const browser = await chromium.launch({ headless: true });

try {
  await runFullLinkSmoke();
  await runCodeSmoke();
} finally {
  await browser.close();
}

async function runFullLinkSmoke() {
  const contextA = await browser.newContext();
  const contextB = await browser.newContext();
  try {
    const pageA = await contextA.newPage();
    await pageA.goto(baseURL, { waitUntil: "domcontentloaded" });
    await pageA.getByRole("button", { name: "创建强密钥房间" }).click();
    await pageA.waitForURL(/\/r\/.+#k=.+/, { timeout: 10000 });
    await enterName(pageA, "Alice");
    const inviteURL = pageA.url();

    const pageB = await contextB.newPage();
    await pageB.goto(inviteURL, { waitUntil: "domcontentloaded" });
    await enterName(pageB, "Bob");

    await assertChatWorks(pageA, pageB);

    const result = {
      mode: "full-link",
      inviteURL,
      pageAMessages: await pageA.locator(".message").allTextContents(),
      pageBMessages: await pageB.locator(".message").allTextContents(),
    };
    console.log(JSON.stringify(result, null, 2));
  } finally {
    await contextA.close();
    await contextB.close();
  }
}

async function runCodeSmoke() {
  const contextA = await browser.newContext();
  const contextB = await browser.newContext();
  try {
    const pageA = await contextA.newPage();
    await pageA.goto(baseURL, { waitUntil: "domcontentloaded" });
    await pageA.getByRole("button", { name: "创建随机码房间" }).click();
    await pageA.waitForURL(/\/r\/[ABCDEFGHJKMNPQRSTUVWXYZ2-9]{4,32}#p=[ABCDEFGHJKMNPQRSTUVWXYZ2-9]{4,32}/, { timeout: 30000 });
    await enterName(pageA, "Alice");
    const inviteURL = pageA.url();
    const code = new URL(inviteURL).pathname.split("/").pop();

    const pageB = await contextB.newPage();
    await pageB.goto(baseURL, { waitUntil: "domcontentloaded" });
    await pageB.getByPlaceholder("输入群聊码").fill(code);
    await pageB.getByRole("button", { name: "用群聊码加入" }).click();
    await pageB.waitForURL(new RegExp(`/r/${code}#p=${code}`), { timeout: 30000 });
    await enterName(pageB, "Bob");

    await assertChatWorks(pageA, pageB);

    const result = {
      mode: "code",
      code,
      inviteURL,
      pageAMessages: await pageA.locator(".message").allTextContents(),
      pageBMessages: await pageB.locator(".message").allTextContents(),
    };
    console.log(JSON.stringify(result, null, 2));
  } finally {
    await contextA.close();
    await contextB.close();
  }
}

async function assertChatWorks(pageA, pageB) {
  await pageA.getByText("已连接").waitFor({ timeout: 10000 });
  await pageB.getByText("已连接").waitFor({ timeout: 10000 });

  await pageA.locator(".members .n-list-item").filter({ hasText: "私发安全码" }).waitFor({ timeout: 10000 });
  await pageB.locator(".members .n-list-item").filter({ hasText: "私发安全码" }).waitFor({ timeout: 10000 });

  await pageA.getByPlaceholder("输入消息").fill("hello from A group");
  await pageA.getByRole("button", { name: "发送群聊" }).click();
  await pageB.getByText("hello from A group").waitFor({ timeout: 10000 });

  await pageA.locator('input[type="file"]').setInputFiles({
    name: "hello.txt",
    mimeType: "text/plain",
    buffer: Buffer.from("encrypted file from A"),
  });
  await pageA.getByRole("button", { name: "发送群聊" }).click();
  await pageB.getByText("hello.txt").waitFor({ timeout: 10000 });

  await pageB.getByPlaceholder("输入消息").fill("hello from B group");
  await pageB.getByRole("button", { name: "发送群聊" }).click();
  await pageA.getByText("hello from B group").waitFor({ timeout: 10000 });

  await pageB.locator('input[type="file"]').setInputFiles({
    name: "pixel.png",
    mimeType: "image/png",
    buffer: tinyPng,
  });
  await pageB.getByRole("button", { name: "发送群聊" }).click();
  await pageA.getByText("pixel.png").waitFor({ timeout: 10000 });

  await pageA.evaluate(async (base64) => {
    const bytes = Uint8Array.from(atob(base64), (char) => char.charCodeAt(0));
    const file = new File([bytes], "paste.png", { type: "image/png" });
    const item = new DataTransfer();
    item.items.add(file);
    const input = document.querySelector('input[placeholder="输入消息"]');
    input.dispatchEvent(new ClipboardEvent("paste", {
      bubbles: true,
      cancelable: true,
      clipboardData: item,
    }));
  }, tinyPng.toString("base64"));
  await pageA.getByText(/pasted-image-|paste\\.png/).waitFor({ timeout: 10000 });
  await pageA.getByRole("button", { name: "发送群聊" }).click();
  await pageB.getByText(/pasted-image-|paste\\.png/).waitFor({ timeout: 10000 });

  await pageA.locator(".members .n-list-item").filter({ hasText: "私发安全码" }).click();
  await pageA.getByPlaceholder("输入消息").fill("private from A to B");
  await pageA.getByRole("button", { name: /^私发给 / }).click();
  await pageB.getByText("private from A to B").waitFor({ timeout: 10000 });
}

async function enterName(page, name) {
  await page.getByPlaceholder("给自己起个名字").fill(name);
  await page.getByRole("button", { name: "进入聊天" }).click();
}
