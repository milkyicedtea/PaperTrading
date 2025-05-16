import "@mantine/core/styles.css"
import {AppShell, MantineProvider} from "@mantine/core"
import { theme } from "@local/theme"
import {RouterProvider} from "@tanstack/react-router";
import {router} from "@local/router.tsx";
import "@local/index.css"

export default function App() {
  return (
    <MantineProvider theme={theme}>
      <AppShell>
        <RouterProvider router={router}/>
      </AppShell>
    </MantineProvider>
  )
}
