/**
 * Color generation utilities
 */

/**
 * Generates a deterministic color from a string
 * Uses hash function to ensure the same string always produces the same color
 * @param str - The string to convert to a color
 * @returns Hexadecimal color code
 */
export const stringToColor = (str: string): string => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  const c = (hash & 0x00FFFFFF).toString(16).toUpperCase();
  return '#' + '00000'.substring(0, 6 - c.length) + c;
};
