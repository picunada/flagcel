import assert from "node:assert/strict";

const { inferPayload, parseSchema, samplePayload, serializeSchema, validatePayload } =
    await import("./context-schema.ts");

const fields = [
    { path: "user.id", type: "string" },
    { path: "user.age", type: "int" },
    { path: "score", type: "double" },
];
assert.deepEqual(parseSchema(serializeSchema(fields)).fields, fields);
assert.equal(parseSchema("bad-path string\nuser.id nope\nuser.id string\nuser.id string").errors.length, 3);

const inferred = inferPayload(
    JSON.stringify({ user: { id: "u1", age: 42, ratio: 0.5, active: true }, tags: [], metadata: {}, missing: null }),
);
assert.deepEqual(inferred.fields, [
    { path: "user.id", type: "string" },
    { path: "user.age", type: "int" },
    { path: "user.ratio", type: "double" },
    { path: "user.active", type: "bool" },
    { path: "tags", type: "list" },
    { path: "metadata", type: "map" },
]);
assert.equal(inferred.warnings.length, 1);

const results = validatePayload(
    JSON.stringify({ user: { id: "u1", age: 42 }, score: 4, extra: true }),
    fields,
);
assert.deepEqual(results.map(({ status, path }) => ({ status, path })), [
    { status: "valid", path: "user.id" },
    { status: "valid", path: "user.age" },
    { status: "valid", path: "score" },
    { status: "extra", path: "extra" },
]);

const sample = JSON.parse(samplePayload(fields));
assert.equal(sample.user.id, "usr_4f2a");
assert.equal(sample.user.age, 42);
