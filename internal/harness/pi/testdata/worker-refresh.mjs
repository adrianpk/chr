// Exercise the pinned Pi credential store with synthetic provider refresh IO.
import { AuthStorage } from "/opt/chr-pi/node_modules/@earendil-works/pi-coding-agent/dist/core/auth-storage.js";

const storage = AuthStorage.create("/private/auth.json");
await storage.modify("fixture", async current => {
  const before = current?.expires;
  if (current?.type !== "oauth" || !Number.isInteger(before) || before < 1 || before > 2 ||
      current.access !== `fixture-access-${before}` || current.refresh !== `fixture-refresh-${before}`) {
    throw new Error("credential handoff lost the previous refresh");
  }
  const next = before + 1;
  return {type: "oauth", access: `fixture-access-${next}`, refresh: `fixture-refresh-${next}`, expires: next};
});
