import { describe, expect, it } from "vitest";
import { routeRoleAccess } from "./App";

describe("routeRoleAccess", () => {
  it("allows teachers to open course document reader routes", () => {
    expect(routeRoleAccess.documentReader).toContain("teacher");
  });
});
