import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";

function renderApp(initial = "/login") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[initial]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function stubFetch(handler: (url: string) => Response | Promise<Response>) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo) => handler(String(input))),
  );
}

describe("App", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders login when bootstrap is complete", async () => {
    stubFetch((url) => {
      if (url.includes("/setup/status")) {
        return new Response(JSON.stringify({ needs_bootstrap: false }), { status: 200 });
      }
      if (url.includes("/auth/me")) {
        return new Response(JSON.stringify({ error: "authentication required" }), { status: 401 });
      }
      return new Response("{}", { status: 404 });
    });

    renderApp("/login");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Sign in" })).toBeInTheDocument();
    });
  });

  it("renders setup when bootstrap is required", async () => {
    stubFetch((url) => {
      if (url.includes("/setup/status")) {
        return new Response(JSON.stringify({ needs_bootstrap: true }), { status: 200 });
      }
      if (url.includes("/auth/me")) {
        return new Response(JSON.stringify({ error: "authentication required" }), { status: 401 });
      }
      return new Response("{}", { status: 404 });
    });

    renderApp("/setup");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Create admin account" })).toBeInTheDocument();
    });
  });
});
