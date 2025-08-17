export const generateComponentColor = (component: string) => {
  // Simple but stable hash → integer
  let hash = 0;
  for (let i = 0; i < component.length; i++) {
    hash = component.charCodeAt(i) + ((hash << 5) - hash);
  }

  // Golden angle (~137.5°) ensures evenly spaced hues
  const goldenAngle = 137.508;
  const hue = (Math.abs(hash) * goldenAngle) % 360;

  // Fix S and L to keep contrast high
  const saturation = 65;
  const lightness = 55;

  return `hsl(${hue}, ${saturation}%, ${lightness}%)`;
};
