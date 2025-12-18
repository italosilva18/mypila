/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
        "./components/**/*.{js,ts,jsx,tsx}",
        "./pages/**/*.{js,ts,jsx,tsx}",
        "./*.{js,ts,jsx,tsx}" // For App.tsx etc if in root
    ],
    theme: {
        extend: {
            fontFamily: {
                sans: ['Nunito', 'sans-serif'],
            },
            colors: {
                paper: {
                    DEFAULT: '#faf8f5',
                    dark: '#f5f3f0',
                },
                ink: {
                    DEFAULT: '#1c1917',
                    light: '#292524',
                    muted: '#78716c',
                    faded: '#a8a29e',
                },
            },
            borderRadius: {
                'xl': '0.75rem',
                '2xl': '1rem',
                '3xl': '1.5rem',
            },
            boxShadow: {
                'soft': '0 2px 15px -3px rgba(0,0,0,0.07)',
                'card': '0 4px 20px -2px rgba(0,0,0,0.1)',
            },
        },
    },
    plugins: [],
}
