import { expect, test } from "bun:test";

import { greet } from "./index.ts";

const cli = new URL("./cli.ts", import.meta.url).pathname;

test("greet", () => {
  expect(greet("Bunny")).toBe("Hello, Bunny!");
  expect(greet()).toBe("Hello, world!");
});

test("cli greets a name", () => {
  const run = Bun.spawnSync(["bun", "run", cli, "Ada"]);
  expect(run.stdout.toString()).toBe("Hello, Ada!\n");
});
