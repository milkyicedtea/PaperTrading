import {Anchor, Container, Divider, Group, Stack, Text} from "@mantine/core";
import {useMediaQuery} from "@mantine/hooks";

export function Footer() {
  const isMobile = useMediaQuery('(max-width: 768px)')

  return (
    <>
      <Divider/>
      <footer style={{paddingTop: '1rem', paddingBottom: '1rem'}}>
        <Container>
          {isMobile ? (
            <Stack align="center" gap="xs">
              <Group gap="sm" justify="center" wrap="wrap">
                <Anchor href="#about" size={'sm'}>About Us</Anchor>
                <Anchor href="#contact" size={'sm'}>Contact</Anchor>
                <Anchor href="#terms" size={'sm'}>Terms</Anchor>
                <Anchor href="#privacy" size={'sm'}>Privacy</Anchor>
              </Group>
              <Text size={'sm'} c={'dimmed'} ta="center">
                &copy; {new Date().getFullYear()} PaperTrading. All rights reserved.
              </Text>
            </Stack>
          ) : (
            <Group justify="space-between" align="center">
              <Text size={'sm'} c={'dimmed'}>
                &copy; {new Date().getFullYear()} PaperTrading. All rights reserved.
              </Text>
              <Group gap="sm">
                <Anchor href="#about" size={'sm'}>About Us</Anchor>
                <Anchor href="#contact" size={'sm'}>Contact</Anchor>
                <Anchor href="#terms" size={'sm'}>Terms</Anchor>
                <Anchor href="#privacy" size={'sm'}>Privacy</Anchor>
              </Group>
            </Group>
          )}
        </Container>
      </footer>
    </>
  )
}