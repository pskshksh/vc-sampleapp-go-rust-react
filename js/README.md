# React + TypeScript + Rsbuild

The frontend for `vc-sampleapp-go-rust-react`. Built with
[Rsbuild](https://rsbuild.dev) (the Rspack-based build tool) via pnpm.

## Scripts

```sh
pnpm dev        # start the dev server on :5173 (proxies /api to goapi)
pnpm build      # production build to dist/
pnpm preview    # preview the production build
pnpm typecheck  # tsc -b
pnpm lint       # oxlint
```

The dev server proxies `/api` to goapi. Point it elsewhere with the `GOAPI_URL`
environment variable (see `rsbuild.config.ts`).
