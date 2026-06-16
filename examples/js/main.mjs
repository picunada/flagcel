import { OpenFeature } from "@openfeature/server-sdk";
import { FlagcelProvider } from "@flagcel/openfeature-server";

const endpoint = env("FLAGCEL_ENDPOINT", "http://localhost:8080/api/v1");
const apiKey = process.env.FLAGCEL_API_KEY ?? "";
const flagKey = env("FLAGCEL_FLAG_KEY", "new-checkout");
const targetingKey = env("TARGETING_KEY", "u_123");

const provider = new FlagcelProvider({
  endpoint,
  apiKey,
});

await OpenFeature.setProviderAndWait(provider);

try {
  const client = OpenFeature.getClient("flagcel-local-example");
  const details = await client.getBooleanDetails(flagKey, false, {
    targetingKey,
    user: {
      id: targetingKey,
      country: env("USER_COUNTRY", "US"),
    },
    request: {
      path: env("REQUEST_PATH", "/checkout"),
    },
  });

  console.log(
    `flag=${flagKey} value=${details.value} reason=${details.reason} variant=${details.variant ?? ""}`,
  );
} finally {
  await OpenFeature.clearProviders();
}

function env(key, fallback) {
  return process.env[key] && process.env[key].length > 0 ? process.env[key] : fallback;
}
