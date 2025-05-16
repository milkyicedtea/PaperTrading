import {createRootRoute, createRouter, HeadContent, Outlet, Scripts} from "@tanstack/react-router";
import {homeRoute} from "@local/pages/routes.ts";
import favicon from "@local/assets/favicon.ico"

export const rootRoute = createRootRoute({
  component: () => (
    <>
      <HeadContent/>
      <Outlet/>
      <Scripts/>
    </>
  ),
  head: () => ({
    meta: [
      {
        name: "viewport",
        content: "minimum-scale=1, initial-scale=1, width=device-width, user-scalable=no"
      }
    ],
    links: [
      {
        rel: "icon",
        href: favicon
      }
    ],
    scripts: [
      {
        src: 'https://www.google-analytics.com/analytics.js',
      },
    ],
  })
})

const routeTree = rootRoute.addChildren([homeRoute])

export const router = createRouter({routeTree})