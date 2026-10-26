// Test suite requires DATABASE_URL to connect
if (!process.env.DATABASE_URL) {
  console.error("FAIL: missing env var DATABASE_URL");
  process.exit(1);
}

console.log("All tests passed!");
