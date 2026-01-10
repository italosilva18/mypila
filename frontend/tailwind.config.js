/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
        "./components/**/*.{js,ts,jsx,tsx}",
        "./pages/**/*.{js,ts,jsx,tsx}",
        "./*.{js,ts,jsx,tsx}"
    ],
    theme: {
        extend: {
            fontFamily: {
                sans: ['Nunito', 'system-ui', 'sans-serif'],
            },
            colors: {
                // Teal/Mint Primary Colors
                primary: {
                    50: 'hsl(168 65% 95%)',
                    100: 'hsl(168 65% 90%)',
                    200: 'hsl(168 65% 80%)',
                    300: 'hsl(168 65% 70%)',
                    400: 'hsl(168 65% 60%)',
                    500: 'hsl(168 65% 55%)',
                    600: 'hsl(168 65% 45%)',
                    700: 'hsl(168 65% 35%)',
                    800: 'hsl(168 65% 25%)',
                    900: 'hsl(168 65% 15%)',
                },
                // Background/Foreground
                background: 'hsl(0 0% 98%)',
                foreground: 'hsl(220 20% 20%)',
                muted: {
                    DEFAULT: 'hsl(220 10% 50%)',
                    foreground: 'hsl(220 10% 70%)',
                },
                // Card
                card: {
                    DEFAULT: 'hsl(0 0% 100%)',
                    foreground: 'hsl(220 20% 20%)',
                },
                // Border
                border: 'hsl(220 10% 90%)',
                // State colors
                success: {
                    DEFAULT: 'hsl(145 60% 50%)',
                    light: 'hsl(145 60% 95%)',
                    dark: 'hsl(145 60% 40%)',
                },
                warning: {
                    DEFAULT: 'hsl(0 55% 70%)',
                    light: 'hsl(0 55% 95%)',
                    dark: 'hsl(0 55% 60%)',
                },
                destructive: {
                    DEFAULT: 'hsl(0 70% 55%)',
                    light: 'hsl(0 70% 95%)',
                    dark: 'hsl(0 70% 45%)',
                },
            },
            borderRadius: {
                DEFAULT: '0.75rem',
                'lg': '0.75rem',
                'xl': '1rem',
                '2xl': '1.25rem',
                '3xl': '1.5rem',
            },
            boxShadow: {
                'soft': '0 2px 15px -3px rgba(0,0,0,0.07)',
                'card': '0 4px 20px -2px rgba(0,0,0,0.1)',
                'hover': '0 8px 30px -4px rgba(0,0,0,0.15)',
            },
            backgroundImage: {
                'gradient-primary': 'linear-gradient(135deg, hsl(168 65% 55%) 0%, hsl(168 50% 65%) 100%)',
                'gradient-header': 'linear-gradient(180deg, hsl(168 65% 50%) 0%, hsl(168 65% 55%) 100%)',
            },
        },
    },
    plugins: [],
}
