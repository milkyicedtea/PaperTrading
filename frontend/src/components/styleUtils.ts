export function getBackgroundColor(colorScheme: "light" | "dark", hovered: boolean) {
  return colorScheme === 'dark'
    ? hovered // dark
      ? "#3b3b3b" // hovered
      : "#2e2e2e" // not hovered
    : hovered // light
      ? "#f8f9fa" // hovered
      : "#ffffff" // not hovered
}

export function getBorder(colorScheme: "light" | "dark") {
  return `.0625rem solid ${colorScheme == "dark" ? "#424242" : "#dee2e6"}`
}