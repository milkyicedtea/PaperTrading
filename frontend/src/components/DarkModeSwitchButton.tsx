import { RxMoon, RxSun } from "react-icons/rx"
import {Container, Tooltip, UnstyledButton, useComputedColorScheme, useMantineColorScheme} from "@mantine/core"
import {useHover} from "@mantine/hooks"
import {getBackgroundColor, getBorder} from "@local/components/styleUtils"

export function DarkModeSwitchButton() {
  const {hovered, ref} = useHover();
  const colorScheme = useComputedColorScheme();
  const {setColorScheme} = useMantineColorScheme()

  async function toggleColorScheme() {
    setColorScheme(colorScheme === "dark" ? "light" : "dark");
  }

  return (
    <>
      <Tooltip label={`Disable ${colorScheme.charAt(0).toUpperCase()+colorScheme.slice(1)} mode`} openDelay={300}>
        <Container style={{display: "flex", placeSelf: "center", marginRight: "1.25rem"}}>
          <UnstyledButton
            ref={ref}
            style={{
              display: "flex",
              borderRadius: ".5rem",
              border: getBorder(colorScheme),
              backgroundColor: getBackgroundColor(colorScheme, hovered),
              aspectRatio: 1,
              width: "2.125rem",
              justifyContent: "center",
              alignItems: "center",
            }}
            onClick={toggleColorScheme}>
             {colorScheme === "dark" ? <RxMoon/> : <RxSun/>}
          </UnstyledButton>
        </Container>
      </Tooltip>

    </>

  )
}