import { assertEquals } from "@std/assert";
import { greet } from "./greet.ts";

Deno.test("greet", () => {
  assertEquals(greet("Deno"), "Hello, Deno!");
  assertEquals(greet(), "Hello, world!");
});
