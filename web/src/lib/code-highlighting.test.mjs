import assert from "node:assert/strict";

const { formatJSON, highlightCode } = await import("./code-highlighting.ts");

assert.equal(formatJSON('{"user":{"id":"u1"},"active":true}'), `{
  "user": {
    "id": "u1"
  },
  "active": true
}`);
assert.equal(formatJSON("{"), "{");
assert.match(highlightCode("user.id string", "schema"), /color:var\(--color-valid\)/);
assert.match(highlightCode('{"count":42,"ok":true}', "json"), /color:var\(--color-cel-warning\)/);
assert.doesNotMatch(highlightCode('{"value":"<script>"}', "json"), /<script>/);
