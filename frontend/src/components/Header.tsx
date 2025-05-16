import {Button, Box, Divider, Group, Title, Burger, Drawer, Stack} from "@mantine/core"
import {DarkModeSwitchButton} from "@local/components/DarkModeSwitchButton"
import {useNavigate} from "@tanstack/react-router";
import {useDisclosure, useMediaQuery} from "@mantine/hooks";

export function Header() {
  const navigate = useNavigate()
  const [opened, { toggle: toggleDrawer, close: closeDrawer }] = useDisclosure(false)
  const isMobile = useMediaQuery('(max-width: 768px)')

  const navigateAndCloseDrawer = (to: string) => {
    void navigate({ to })
    closeDrawer()
  }

  const navItems = (
    <>
      <Button variant={'subtle'} onClick={() => navigateAndCloseDrawer('#features')}>Features</Button>
      <Button variant={'subtle'} onClick={() => navigateAndCloseDrawer('#how-it-works')}>How it Works</Button>
      <Button variant={'subtle'} onClick={() => navigateAndCloseDrawer('#pricing')}>Pricing</Button>
      <Button variant={'subtle'} onClick={() => navigateAndCloseDrawer('#learn')}>Learn</Button>
    </>
  )

  const authItems = (
    <>
      <Button  variant="outline" onClick={() => {closeDrawer()}}>Login</Button>
      <Button  onClick={() => {closeDrawer()}}>Sign Up</Button>
    </>
  );

  return (
    <>
      <header style={{
        display: 'flex',
        alignItems: 'center',
        height: 56,
        width: '100%',
        boxSizing: 'border-box',
        position: 'sticky',
        top: 0,
        zIndex: 100,
        background: 'var(--mantine-color-body)',
        borderBottom: '1px solid light-dark(var(--mantine-color-gray-3), var(--mantine-color-dark-4))'
      }}>
        <Box
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            width: '100%',
            paddingLeft: '1rem',
            paddingRight: '1rem',
            boxSizing: 'border-box',
          }}
        >
          <Title order={2}>
            📈 PaperTrading
          </Title>

          {isMobile ? (
            <Group gap="sm">
              <DarkModeSwitchButton />
              <Burger opened={opened} onClick={toggleDrawer} aria-label="Toggle navigation" />
            </Group>
          ) : (
            <>
              <Group gap="xs" wrap="nowrap" visibleFrom="sm">
                {navItems}
              </Group>
              <Group gap="sm" wrap="nowrap" visibleFrom="sm">
                {authItems}
                <DarkModeSwitchButton />
              </Group>
            </>
          )}
        </Box>

      </header>

      {isMobile && (
        <Drawer
          opened={opened}
          onClose={closeDrawer}
          title="Menu"
          padding="md"
          size="md"
          position="right"
        >
          <Stack gap="md">
            {navItems}
            <Divider />
            {authItems}
            {/* could also place DarkModeSwitchButton here for mobile*/}
          </Stack>
        </Drawer>
      )}
    </>
  )
}