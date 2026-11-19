import { assertEquals } from "@std/assert";
import { run } from "./main.ts";

Deno.test("run greets the world by default", () => {
  assertEquals(run([]), "Hello, world!");
});

Deno.test("run greets a name", () => {
  assertEquals(run(["Ada"]), "Hello, Ada!");
});
