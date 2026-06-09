import { ErrorBoundary, LocationProvider, Route, Router } from "preact-iso";
import { locationStub } from "preact-iso/prerender";
import Home from "./Home";
import User from "./User";

type AppProps = {
  url?: string;
};

function NotFound() {
  return (
    <main class="min-h-screen bg-zinc-950 px-6 py-12 text-zinc-100">
      <section class="mx-auto max-w-3xl rounded-2xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
        <p class="mb-3 text-sm font-medium uppercase tracking-wide text-zinc-400">
          Not found
        </p>

        <h1 class="text-4xl font-bold tracking-tight">
          This route does not exist.
        </h1>

        <a class="mt-8 inline-block text-zinc-300 underline" href="/">
          Back home
        </a>
      </section>
    </main>
  );
}

export default function App(props: AppProps) {
  if (typeof window === "undefined") {
    locationStub(props.url ?? "/");
  }

  return (
    <LocationProvider>
      <ErrorBoundary>
        <Router>
          <Route path="/" component={Home} />
          <Route path="/users/:id" component={User} />
          <Route default component={NotFound} />
        </Router>
      </ErrorBoundary>
    </LocationProvider>
  );
}
