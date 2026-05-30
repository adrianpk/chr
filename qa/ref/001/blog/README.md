# Blog

Small headless blog service used as a ChrysoCoord QA reference project.

From the parent `chc` repository:

```sh
make qa-reset-blog
make qa-test-blog
```

The initial baseline intentionally has one failing test:
`GET /posts/{slug}` returns `500` for an unknown slug instead of `404`.

That failure seeds the first manual agent-work ticket in a later slice.
