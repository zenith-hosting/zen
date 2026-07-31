# Zen Starter

```bash
git clone https://github.com/zenith-hosting/zen-starter my-app
cd my-app
go mod edit -module example.com/my-app
pnpm tidy
pnpm dev
```

Use `pnpm build` to create production artifacts and `pnpm start` to run them. The `package.json` scripts contain the complete workflow.
