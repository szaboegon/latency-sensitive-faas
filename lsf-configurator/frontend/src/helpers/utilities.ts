export const generateComponentColor = (component: string) => {
  let hash = 5381;
  for (let i = 0; i < component.length; i++) {
    hash = (hash * 33) ^ component.charCodeAt(i);
  }

  // Ensure positive number and map to hue
  const hue = Math.abs(hash) % 360;

  return `hsl(${hue}, 70%, 80%)`;
};