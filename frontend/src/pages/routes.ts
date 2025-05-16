import {createRoute} from "@tanstack/react-router"
import {rootRoute} from "@local/router.tsx"

export const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  head: () => ({
    meta: [
      {
        name: "description",
        content: "Practice your trading with PaperTrading",
      },
      {
        title: "PaperTrading | Trading practice"
      }
    ],
  })
}).lazy(() => import("@local/pages/Home.tsx").then((d) => d.Route))